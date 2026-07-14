package health

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/raifmirza/consistent-hash/internal/backend"
	"github.com/raifmirza/consistent-hash/internal/node"
)

type sequenceProber struct {
	mu      sync.Mutex
	results []error
}

func (p *sequenceProber) Probe(context.Context, *node.Node) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.results) == 0 {
		return nil
	}

	err := p.results[0]
	p.results = p.results[1:]
	return err
}

type trackingRing struct {
	mu      sync.Mutex
	added   []string
	removed []string
}

func (r *trackingRing) AddNode(n *node.Node) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.added = append(r.added, n.ID)
}

func (r *trackingRing) RemoveNode(n *node.Node) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.removed = append(r.removed, n.ID)
}

func (r *trackingRing) addedIDs() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.added...)
}

func (r *trackingRing) removedIDs() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.removed...)
}

func runCheckerFor(t *testing.T, hc *HealthChecker, duration time.Duration) context.CancelFunc {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	go hc.Start(ctx)
	time.Sleep(duration)
	return cancel
}

func TestHealthChecker_AddsHealthyBackendAfterSuccessThreshold(t *testing.T) {
	registry := backend.NewRegistry()
	n := &node.Node{ID: "server-1", Address: "http://localhost:8081"}
	if _, err := registry.Register(n); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	ring := &trackingRing{}
	prober := &sequenceProber{}

	hc := NewHealthChecker(registry, ring, 20*time.Millisecond, prober, 3, 2)
	cancel := runCheckerFor(t, hc, 100*time.Millisecond)
	cancel()

	if added := ring.addedIDs(); len(added) != 1 || added[0] != "server-1" {
		t.Fatalf("expected backend to be added after 2 successes, got %v", added)
	}
}

func TestHealthChecker_RemovesUnhealthyBackendAfterFailureThreshold(t *testing.T) {
	registry := backend.NewRegistry()
	n := &node.Node{ID: "server-1", Address: "http://localhost:8081"}
	if _, err := registry.Register(n); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	ring := &trackingRing{}
	probeErr := errors.New("probe failed")
	prober := &sequenceProber{
		results: []error{nil, nil, probeErr, probeErr, probeErr, probeErr, probeErr},
	}

	hc := NewHealthChecker(registry, ring, 20*time.Millisecond, prober, 3, 2)
	cancel := runCheckerFor(t, hc, 200*time.Millisecond)
	cancel()

	if removed := ring.removedIDs(); len(removed) != 1 || removed[0] != "server-1" {
		t.Fatalf("expected backend to be removed after 3 failures, got %v", removed)
	}
}

func TestHealthChecker_DoesNotReaddAlreadyHealthyBackend(t *testing.T) {
	registry := backend.NewRegistry()
	n := &node.Node{ID: "server-1", Address: "http://localhost:8081"}
	if _, err := registry.Register(n); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	ring := &trackingRing{}
	prober := &sequenceProber{}

	hc := NewHealthChecker(registry, ring, 20*time.Millisecond, prober, 3, 2)
	cancel := runCheckerFor(t, hc, 150*time.Millisecond)
	cancel()

	if added := ring.addedIDs(); len(added) != 1 {
		t.Fatalf("expected backend to be added once, got %v", added)
	}
}
