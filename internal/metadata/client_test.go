package metadata

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
