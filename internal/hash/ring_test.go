package hash

import (
	"testing"

	"github.com/raifmirza/consistent-hash/internal/node"
)

func TestGetBackendID_EmptyRing(t *testing.T) {
	ring := NewHashRing(10)

	_, err := ring.GetBackendID("user-1")
	if err != ErrEmptyRing {
		t.Fatalf("expected ErrEmptyRing, got %v", err)
	}
}

func TestGetBackendID_ConsistentRouting(t *testing.T) {
	ring := NewHashRing(10)
	ring.AddNode(&node.Node{ID: "a", Address: "http://localhost:8081"})
	ring.AddNode(&node.Node{ID: "b", Address: "http://localhost:8082"})
	ring.AddNode(&node.Node{ID: "c", Address: "http://localhost:8083"})

	first, err := ring.GetBackendID("user-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 5; i++ {
		backendID, err := ring.GetBackendID("user-42")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if backendID != first {
			t.Fatalf("expected stable routing, got %q then %q", first, backendID)
		}
	}
}

func TestAddNode_IsIdempotent(t *testing.T) {
	ring := NewHashRing(3)
	n := &node.Node{ID: "server-1", Address: "http://localhost:8081"}

	ring.AddNode(n)
	ring.AddNode(n)

	backendID, err := ring.GetBackendID("any-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backendID != "server-1" {
		t.Fatalf("expected single node in ring, got %q", backendID)
	}
}

func TestRemoveNode(t *testing.T) {
	ring := NewHashRing(10)
	n1 := &node.Node{ID: "a", Address: "http://localhost:8081"}
	n2 := &node.Node{ID: "b", Address: "http://localhost:8082"}

	ring.AddNode(n1)
	ring.AddNode(n2)
	ring.RemoveNode(n1)

	backendID, err := ring.GetBackendID("user-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backendID != "b" {
		t.Fatalf("expected remaining node b, got %q", backendID)
	}
}

func TestRemoveNode_NotPresent(t *testing.T) {
	ring := NewHashRing(10)
	ring.AddNode(&node.Node{ID: "a", Address: "http://localhost:8081"})

	ring.RemoveNode(&node.Node{ID: "missing", Address: "http://localhost:9999"})

	backendID, err := ring.GetBackendID("user-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backendID != "a" {
		t.Fatalf("expected node a to remain, got %q", backendID)
	}
}

func TestGetBackendID_WrapsAroundRing(t *testing.T) {
	ring := NewHashRing(1)
	ring.AddNode(&node.Node{ID: "only", Address: "http://localhost:8081"})

	backendID, err := ring.GetBackendID("wrap-around-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backendID != "only" {
		t.Fatalf("expected wrap-around lookup to return only node, got %q", backendID)
	}
}
