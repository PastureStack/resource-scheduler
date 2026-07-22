package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PastureStack/resource-scheduler/internal/metadata"
)

type failingMetadataClient struct {
	metadata.Client
}

func (failingMetadataClient) GetVersion() (string, error) {
	return "", errors.New("metadata unavailable")
}

func TestControlPlanePingURL(t *testing.T) {
	tests := map[string]string{
		"http://server:8080/v1":       "http://server:8080/ping",
		"http://server:8080/v2-beta":  "http://server:8080/ping",
		"https://server/v3/":          "https://server/ping",
		"https://server/base/v2-beta": "https://server/base/ping",
	}
	for input, expected := range tests {
		actual, err := controlPlanePingURL(input)
		if err != nil || actual != expected {
			t.Errorf("controlPlanePingURL(%q) = %q, %v; want %q", input, actual, err, expected)
		}
	}
}

func TestControlPlanePingURLRejectsUnsafeValues(t *testing.T) {
	for _, input := range []string{"", "file:///tmp/server", "http://user:password@server/v1", "http://server/v1?token=value"} {
		if _, err := controlPlanePingURL(input); err == nil {
			t.Errorf("expected %q to be rejected", input)
		}
	}
}

func TestHealthCheckIsLivenessAndReadinessChecksDependencies(t *testing.T) {
	controlPlane := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Error(response, "unavailable", http.StatusServiceUnavailable)
	}))
	defer controlPlane.Close()

	handler, err := newHealthHandler(failingMetadataClient{}, controlPlane.URL+"/v1", controlPlane.Client())
	if err != nil {
		t.Fatalf("newHealthHandler returned an error: %v", err)
	}

	liveness := httptest.NewRecorder()
	handler.ServeHTTP(liveness, httptest.NewRequest(http.MethodGet, "/healthcheck", nil))
	if liveness.Code != http.StatusOK || liveness.Body.String() != "ok" {
		t.Fatalf("liveness = HTTP %d %q; want HTTP 200 ok", liveness.Code, liveness.Body.String())
	}

	readiness := httptest.NewRecorder()
	handler.ServeHTTP(readiness, httptest.NewRequest(http.MethodGet, "/readiness", nil))
	if readiness.Code != http.StatusServiceUnavailable {
		t.Fatalf("readiness = HTTP %d; want HTTP 503 while dependencies are unavailable", readiness.Code)
	}
}

func TestSupervisorReconnectsAfterCleanDisconnect(t *testing.T) {
	runs := 0
	waits := 0
	superviseComponent("test component", func() error {
		runs++
		return nil
	}, func(delay time.Duration) bool {
		waits++
		if delay != componentRetryDelay {
			t.Fatalf("retry delay = %v; want %v", delay, componentRetryDelay)
		}
		return waits < 2
	})

	if runs != 2 {
		t.Fatalf("component ran %d times; want 2", runs)
	}
}

func TestRunComponentSafelyConvertsPanicToError(t *testing.T) {
	err := runComponentSafely(func() error {
		panic("metadata disconnected")
	})
	if err == nil || err.Error() != "component panicked: metadata disconnected" {
		t.Fatalf("panic conversion = %v; want component panic error", err)
	}
}
