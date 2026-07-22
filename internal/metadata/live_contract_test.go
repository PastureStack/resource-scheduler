package metadata

import (
	"os"
	"testing"
)

func TestLiveMetadataContract(t *testing.T) {
	rawURL := os.Getenv("PASTURESTACK_METADATA_TEST_URL")
	if rawURL == "" {
		t.Skip("PASTURESTACK_METADATA_TEST_URL is not set")
	}
	client := NewClient(rawURL)
	version, err := client.GetVersion()
	if err != nil || version == "" {
		t.Fatalf("metadata version contract failed: %v", err)
	}
	if _, err := client.GetHosts(); err != nil {
		t.Fatalf("hosts contract failed: %v", err)
	}
	if _, err := client.GetContainers(); err != nil {
		t.Fatalf("containers contract failed: %v", err)
	}
}
