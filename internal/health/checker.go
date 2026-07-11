package health

import (
	"context"
	"sync"
	"time"

	"github.com/raifmirza/consistent-hash/internal/backend"
	"github.com/raifmirza/consistent-hash/internal/hash"
	"github.com/raifmirza/consistent-hash/internal/node"
)

type BackendProvider interface {
	List() []*backend.Backend
}

type HealthChecker struct {
	mutex sync.Mutex

	provider BackendProvider
	ring     hash.RingManager
	prober   Prober

	status map[string]*NodeStatus

	interval time.Duration

	failureThreshold int
	successThreshold int
}

func NewHealthChecker(provider BackendProvider, ring hash.RingManager, interval time.Duration, prober Prober, failureThreshold int, successThreshold int) *HealthChecker {
	status := make(map[string]*NodeStatus)

	return &HealthChecker{
		provider:         provider,
		ring:             ring,
		prober:           prober,
		interval:         interval,
		status:           status,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
	}
}

func (hc *HealthChecker) Start(ctx context.Context) {
	hc.checkAll(ctx)

	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.checkAll(ctx)
		}
	}
}

func (hc *HealthChecker) checkAndUpdate(parentCtx context.Context, n *node.Node) {
	ctx, cancel := context.WithTimeout(
		parentCtx,
		3*time.Second,
	)
	defer cancel()

	err := hc.prober.Probe(ctx, n)

	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	status, ok := hc.status[n.ID]
	if !ok {
		status = &NodeStatus{
			State: StateUnknown,
		}
		hc.status[n.ID] = status
	}

	if err == nil {
		status.ConsecutiveFailures = 0
		status.ConsecutiveSuccesses++
		status.LastError = nil

		if status.State != StateHealthy && status.ConsecutiveSuccesses >= hc.successThreshold {
			status.State = StateHealthy
			status.ConsecutiveSuccesses = 0

			hc.ring.AddNode(n)

		}
		return
	}

	status.ConsecutiveSuccesses = 0
	status.ConsecutiveFailures++
	status.LastError = err

	if status.State != StateUnhealthy && status.ConsecutiveFailures >= hc.failureThreshold {
		status.State = StateUnhealthy
		status.ConsecutiveFailures = 0

		hc.ring.RemoveNode(n)

	}
}

func (hc *HealthChecker) checkAll(ctx context.Context) {
	backends := hc.provider.List()
	wg := sync.WaitGroup{}
	for _, b := range backends {
		wg.Add(1)
		go func(b *backend.Backend) {
			defer wg.Done()
			hc.checkAndUpdate(ctx, b.Node)
		}(b)
	}
	wg.Wait()
}
