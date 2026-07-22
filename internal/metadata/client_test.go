package metadata

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

type testTimeoutError struct{}

func (testTimeoutError) Error() string   { return "temporary timeout" }
func (testTimeoutError) Timeout() bool   { return true }
func (testTimeoutError) Temporary() bool { return true }

func TestClientReadsSchedulerMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Accept") != "application/json" {
			t.Errorf("unexpected Accept header %q", request.Header.Get("Accept"))
		}
		switch request.URL.Path {
		case "/2016-07-29/version":
			fmt.Fprint(response, `"revision-1"`)
		case "/2016-07-29/hosts":
			fmt.Fprint(response, `[{"uuid":"host-1","memory":2048,"milli_cpu":2000}]`)
		case "/2016-07-29/containers":
			fmt.Fprint(response, `[{"uuid":"container-1","host_uuid":"host-1","state":"running","ports":["0.0.0.0:8080:80/tcp"]}]`)
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL + "/2016-07-29")
	version, err := client.GetVersion()
	if err != nil || version != "revision-1" {
		t.Fatalf("GetVersion = %q, %v", version, err)
	}
	hosts, err := client.GetHosts()
	if err != nil || len(hosts) != 1 || hosts[0].UUID != "host-1" {
		t.Fatalf("GetHosts = %#v, %v", hosts, err)
	}
	containers, err := client.GetContainers()
	if err != nil || len(containers) != 1 || containers[0].HostUUID != "host-1" {
		t.Fatalf("GetContainers = %#v, %v", containers, err)
	}
}

func TestClientRejectsUnsafeMetadataURLs(t *testing.T) {
	for _, rawURL := range []string{
		"file:///tmp/metadata",
		"http://user:password@example.test/metadata",
		"http://example.test/metadata?token=value",
		"http://example.test/metadata#fragment",
	} {
		if _, err := NewClient(rawURL).GetVersion(); err == nil {
			t.Errorf("expected %q to be rejected", rawURL)
		}
	}
}

func TestClientBlocksCrossOriginRedirects(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		fmt.Fprint(response, `"unexpected"`)
	}))
	defer target.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Redirect(response, request, target.URL, http.StatusFound)
	}))
	defer origin.Close()

	_, err := NewClient(origin.URL).GetVersion()
	if err == nil || !strings.Contains(err.Error(), "redirect changed origin") {
		t.Fatalf("unexpected redirect error: %v", err)
	}
}

func TestClientDoesNotExposeResponseBodiesInErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.WriteHeader(http.StatusBadGateway)
		fmt.Fprint(response, "sensitive-response-value")
	}))
	defer server.Close()

	_, err := NewClient(server.URL).GetVersion()
	if err == nil || strings.Contains(err.Error(), "sensitive-response-value") {
		t.Fatalf("unexpected response error: %v", err)
	}
}

func TestWaitVersionRetriesTransientTimeout(t *testing.T) {
	client, err := newHTTPClient("http://metadata.test/2016-07-29")
	if err != nil {
		t.Fatal(err)
	}
	requests := 0
	client.client.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		if requests == 1 {
			return nil, &url.Error{Op: request.Method, URL: request.URL.String(), Err: testTimeoutError{}}
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`"revision-2"`)),
			Request:    request,
		}, nil
	})

	version, err := client.waitVersion(5, "revision-1")
	if err != nil {
		t.Fatal(err)
	}
	if version != "revision-2" {
		t.Fatalf("waitVersion = %q", version)
	}
	if requests != 2 {
		t.Fatalf("request count = %d, want 2", requests)
	}

	var timeout interface {
		Timeout() bool
	}
	if !errors.As(&url.Error{Err: testTimeoutError{}}, &timeout) || !timeout.Timeout() {
		t.Fatal("test timeout did not preserve timeout semantics")
	}
}
