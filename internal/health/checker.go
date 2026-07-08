package health

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/raifmirza/consistent-hash/internal/hash"
	"github.com/raifmirza/consistent-hash/internal/node"
)

type HealthChecker struct {
	mutex sync.Mutex

	ring *hash.HashRing

	prober Prober

	interval time.Duration

	nodes map[string]*node.Node

	status map[string]*NodeStatus

	failureThreshold int

	successThreshold int
}

func NewHealthChecker(ring *hash.HashRing, interval time.Duration, prober Prober, nodes []*node.Node, failureThreshold int, successThreshold int) *HealthChecker {
	nodeMap := make(map[string]*node.Node)
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

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

func (hc *HealthChecker) Start() {

	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		<-ticker.C
		hc.checkAll()
	}
}

func (hc *HealthChecker) checkAndUpdate(n *node.Node) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
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

		if status.State != StateHealthy && status.ConsecutiveFailures >= hc.failureThreshold {
			status.State = StateHealthy
			status.ConsecutiveSuccesses = 0

			hc.ring.AddNode(n)

			log.Printf("%s recovered", n.ID)
		}
		return
	}

	status.ConsecutiveSuccesses = 0
	status.ConsecutiveFailures++

	if status.State != StateUnhealthy && status.ConsecutiveFailures >= hc.failureThreshold {
		status.State = StateUnhealthy
		status.ConsecutiveFailures = 0

		hc.ring.RemoveNode(n)

		log.Printf("%s became unhealthy", n.ID)
	}
}

func (hc *HealthChecker) checkAll() {
	wg := sync.WaitGroup{}
	for _, n := range hc.nodes {
		wg.Add(1)
		go func(n *node.Node) {
			defer wg.Done()
			hc.checkAndUpdate(n)
		}(n)
	}
	wg.Wait()
}
