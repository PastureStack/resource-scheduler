package main

import "testing"

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
