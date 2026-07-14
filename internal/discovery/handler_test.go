package discovery

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/raifmirza/consistent-hash/internal/backend"
	"github.com/raifmirza/consistent-hash/internal/hash"
	"github.com/raifmirza/consistent-hash/internal/node"
)

type trackingRing struct {
	removed []*node.Node
}

func (r *trackingRing) AddNode(*node.Node) {}

func (r *trackingRing) RemoveNode(n *node.Node) {
	r.removed = append(r.removed, n)
}

func TestHandler_Register(t *testing.T) {
	registry := backend.NewRegistry()
	handler := NewHandler(registry, hash.NewHashRing(10))

	body := `{"id":"server-1","address":"http://localhost:8081"}`
	req := httptest.NewRequest(http.MethodPost, "/backends", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
	if !registry.Exists("server-1") {
		t.Fatal("expected backend to be registered")
	}
}

func TestHandler_RegisterInvalidJSON(t *testing.T) {
	handler := NewHandler(backend.NewRegistry(), hash.NewHashRing(10))

	req := httptest.NewRequest(http.MethodPost, "/backends", strings.NewReader("not-json"))
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestHandler_RegisterDuplicate(t *testing.T) {
	registry := backend.NewRegistry()
	handler := NewHandler(registry, hash.NewHashRing(10))

	body := `{"id":"server-1","address":"http://localhost:8081"}`
	req := httptest.NewRequest(http.MethodPost, "/backends", bytes.NewBufferString(body))
	handler.Register(httptest.NewRecorder(), req)

	req = httptest.NewRequest(http.MethodPost, "/backends", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	handler.Register(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", rec.Code)
	}
}

func TestHandler_Delete(t *testing.T) {
	registry := backend.NewRegistry()
	ring := &trackingRing{}
	handler := NewHandler(registry, ring)

	if _, err := registry.Register(&node.Node{
		ID:      "server-1",
		Address: "http://localhost:8081",
	}); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/backends/server-1", nil)
	req.SetPathValue("id", "server-1")
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
	if registry.Exists("server-1") {
		t.Fatal("expected backend to be removed from registry")
	}
	if len(ring.removed) != 1 || ring.removed[0].ID != "server-1" {
		t.Fatal("expected backend to be removed from ring")
	}
}

func TestHandler_DeleteNotFound(t *testing.T) {
	handler := NewHandler(backend.NewRegistry(), hash.NewHashRing(10))

	req := httptest.NewRequest(http.MethodDelete, "/backends/missing", nil)
	req.SetPathValue("id", "missing")
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}
