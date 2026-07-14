package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/raifmirza/consistent-hash/internal/node"
)

func TestHTTPProber_ProbeSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("expected /health path, got %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	prober := NewHTTPProber(2 * time.Second)
	err := prober.Probe(context.Background(), &node.Node{
		ID:      "server-1",
		Address: server.URL,
	})
	if err != nil {
		t.Fatalf("expected successful probe, got %v", err)
	}
}

func TestHTTPProber_ProbeNonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	prober := NewHTTPProber(2 * time.Second)
	err := prober.Probe(context.Background(), &node.Node{
		ID:      "server-1",
		Address: server.URL,
	})
	if err == nil {
		t.Fatal("expected probe failure for non-200 response")
	}
}

func TestHTTPProber_ProbeUnreachable(t *testing.T) {
	prober := NewHTTPProber(100 * time.Millisecond)
	err := prober.Probe(context.Background(), &node.Node{
		ID:      "server-1",
		Address: "http://127.0.0.1:1",
	})
	if err == nil {
		t.Fatal("expected probe failure for unreachable backend")
	}
}
