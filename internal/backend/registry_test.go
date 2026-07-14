package backend

import (
	"testing"

	"github.com/raifmirza/consistent-hash/internal/node"
)

func TestRegistry_RegisterAndLookup(t *testing.T) {
	registry := NewRegistry()
	n := &node.Node{ID: "server-1", Address: "http://localhost:8081"}

	backend, err := registry.Register(n)
	if err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if backend.Node != n {
		t.Fatal("expected registered backend to reference node")
	}

	found, err := registry.Lookup("server-1")
	if err != nil {
		t.Fatalf("unexpected lookup error: %v", err)
	}
	if found.Node.ID != "server-1" {
		t.Fatalf("expected server-1, got %q", found.Node.ID)
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	registry := NewRegistry()
	n := &node.Node{ID: "server-1", Address: "http://localhost:8081"}

	if _, err := registry.Register(n); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if _, err := registry.Register(n); err == nil {
		t.Fatal("expected duplicate registration to fail")
	}
}

func TestRegistry_RegisterInvalidAddress(t *testing.T) {
	registry := NewRegistry()
	n := &node.Node{ID: "server-1", Address: "://bad-url"}

	if _, err := registry.Register(n); err == nil {
		t.Fatal("expected invalid address to fail")
	}
}

func TestRegistry_Remove(t *testing.T) {
	registry := NewRegistry()
	n := &node.Node{ID: "server-1", Address: "http://localhost:8081"}

	if _, err := registry.Register(n); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if err := registry.Remove("server-1"); err != nil {
		t.Fatalf("unexpected remove error: %v", err)
	}
	if registry.Exists("server-1") {
		t.Fatal("expected backend to be removed")
	}
}

func TestRegistry_RemoveNotFound(t *testing.T) {
	registry := NewRegistry()

	if err := registry.Remove("missing"); err == nil {
		t.Fatal("expected remove of missing backend to fail")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	for _, spec := range []struct {
		id   string
		port string
	}{
		{"a", "8081"},
		{"b", "8082"},
		{"c", "8083"},
	} {
		if _, err := registry.Register(&node.Node{
			ID:      spec.id,
			Address: "http://localhost:" + spec.port,
		}); err != nil {
			t.Fatalf("unexpected register error: %v", err)
		}
	}

	backends := registry.List()
	if len(backends) != 3 {
		t.Fatalf("expected 3 backends, got %d", len(backends))
	}
}
