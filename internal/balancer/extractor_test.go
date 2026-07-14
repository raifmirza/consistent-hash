package balancer

import (
	"net/http"
	"testing"
)

func TestIpExtractor_UsesForwardedForHeader(t *testing.T) {
	extractor := IpExtractor{}
	req := httptestRequestWithRemoteAddr("127.0.0.1:1234")
	req.Header.Set("X-Forwarded-For", "10.0.0.5")

	key, err := extractor.Extract(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "10.0.0.5" {
		t.Fatalf("expected forwarded IP, got %q", key)
	}
}

func TestIpExtractor_UsesRemoteAddr(t *testing.T) {
	extractor := IpExtractor{}
	req := httptestRequestWithRemoteAddr("192.168.1.10:4321")

	key, err := extractor.Extract(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "192.168.1.10" {
		t.Fatalf("expected remote host, got %q", key)
	}
}

func TestIpExtractor_InvalidRemoteAddr(t *testing.T) {
	extractor := IpExtractor{}
	req := httptestRequestWithRemoteAddr("invalid-address")

	_, err := extractor.Extract(req)
	if err != ErrNoKey {
		t.Fatalf("expected ErrNoKey, got %v", err)
	}
}

func httptestRequestWithRemoteAddr(remoteAddr string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	if err != nil {
		panic(err)
	}
	req.RemoteAddr = remoteAddr
	return req
}
