package balancer

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/raifmirza/consistent-hash/internal/backend"
	"github.com/raifmirza/consistent-hash/internal/hash"
	"github.com/raifmirza/consistent-hash/internal/node"
)

type stubExtractor struct {
	key string
	err error
}

func (s stubExtractor) Extract(*http.Request) (string, error) {
	return s.key, s.err
}

func TestLoadBalancer_ServeHTTP_ProxiesToBackend(t *testing.T) {
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend-ok"))
	}))
	defer backendServer.Close()

	registry := backend.NewRegistry()
	if _, err := registry.Register(&node.Node{
		ID:      "server-1",
		Address: backendServer.URL,
	}); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	ring := hash.NewHashRing(10)
	ring.AddNode(&node.Node{ID: "server-1", Address: backendServer.URL})

	lb := NewLoadBalancer(ring, registry, stubExtractor{key: "user-42"})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/hello", nil)
	rec := httptest.NewRecorder()

	lb.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "backend-ok" {
		t.Fatalf("expected proxied response, got %q", rec.Body.String())
	}
}

func TestLoadBalancer_ServeHTTP_ExtractorError(t *testing.T) {
	lb := NewLoadBalancer(
		hash.NewHashRing(10),
		backend.NewRegistry(),
		stubExtractor{err: ErrNoKey},
	)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	rec := httptest.NewRecorder()

	lb.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestLoadBalancer_ServeHTTP_EmptyRing(t *testing.T) {
	lb := NewLoadBalancer(
		hash.NewHashRing(10),
		backend.NewRegistry(),
		stubExtractor{key: "user-42"},
	)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	rec := httptest.NewRecorder()

	lb.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rec.Code)
	}
}

func TestLoadBalancer_ServeHTTP_BackendNotFound(t *testing.T) {
	ring := hash.NewHashRing(10)
	ring.AddNode(&node.Node{ID: "ghost", Address: "http://localhost:8081"})

	lb := NewLoadBalancer(
		ring,
		backend.NewRegistry(),
		stubExtractor{key: "user-42"},
	)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	rec := httptest.NewRecorder()

	lb.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rec.Code)
	}
}

func TestLoadBalancer_ServeHTTP_RouterError(t *testing.T) {
	lb := NewLoadBalancer(
		stubRouter{err: errors.New("router failed")},
		backend.NewRegistry(),
		stubExtractor{key: "user-42"},
	)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	rec := httptest.NewRecorder()

	lb.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rec.Code)
	}
}

type stubRouter struct {
	backendID string
	err       error
}

func (s stubRouter) GetBackendID(string) (string, error) {
	return s.backendID, s.err
}
