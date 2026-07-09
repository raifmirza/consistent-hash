package health

import (
	"context"
	"sync"
	"time"

	"github.com/raifmirza/consistent-hash/internal/hash"
	"github.com/raifmirza/consistent-hash/internal/node"
)

type HealthChecker struct {
	mutex sync.Mutex

	ring hash.RingManager

	prober Prober

	interval time.Duration

	nodes map[string]*node.Node

	status map[string]*NodeStatus

	failureThreshold int

	successThreshold int
}

func NewHealthChecker(ring hash.RingManager, interval time.Duration, prober Prober, nodes []*node.Node, failureThreshold int, successThreshold int) *HealthChecker {
	nodeMap := make(map[string]*node.Node)
	status := make(map[string]*NodeStatus)

	for _, n := range nodes {
		status[n.ID] = &NodeStatus{State: StateUnknown}
		nodeMap[n.ID] = n
	}

	return &HealthChecker{
		ring:             ring,
		prober:           prober,
		interval:         interval,
		nodes:            nodeMap,
		status:           status,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
	}
}

func (hc *HealthChecker) Start(ctx context.Context) {

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

	status := hc.status[n.ID]

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
	wg := sync.WaitGroup{}
	for _, n := range hc.nodes {
		wg.Add(1)
		go func(n *node.Node) {
			defer wg.Done()
			hc.checkAndUpdate(ctx, n)
		}(n)
	}
	wg.Wait()
}
