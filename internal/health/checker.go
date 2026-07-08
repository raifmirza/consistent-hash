package health

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/raifmirza/consistent-hash/internal/hash"
	"github.com/raifmirza/consistent-hash/internal/node"
)

type HealthChecker struct {
	mutex sync.Mutex

	ring *hash.HashRing

	client *http.Client

	interval time.Duration

	nodes map[string]*node.Node

	status map[string]*NodeStatus

	failureThreshold int

	successThreshold int
}

func NewHealthChecker(ring *hash.HashRing, interval time.Duration, nodes []*node.Node, failureThreshold int, successThreshold int) *HealthChecker {
	nodeMap := make(map[string]*node.Node)
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	status := make(map[string]*NodeStatus)

	for _, n := range nodes {
		status[n.ID] = &NodeStatus{Healthy: true}
		nodeMap[n.ID] = n
	}

	return &HealthChecker{
		ring:             ring,
		client:           &http.Client{Timeout: 2 * time.Second},
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

func (hc *HealthChecker) checkNode(n *node.Node) bool {
	resp, err := hc.client.Get(n.Address + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (hc *HealthChecker) checkAndUpdate(n *node.Node) {
	healthy := hc.checkNode(n)
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	status := hc.status[n.ID]

	if healthy {
		status.ConsecutiveFailures = 0
		status.ConsecutiveSuccesses++

		if !status.Healthy && status.ConsecutiveFailures >= hc.failureThreshold {
			status.Healthy = true
			status.ConsecutiveSuccesses = 0

			hc.ring.AddNode(n)

			log.Printf("%s recovered", n.ID)
		}
		return
	}

	status.ConsecutiveSuccesses = 0
	status.ConsecutiveFailures++

	if status.Healthy && status.ConsecutiveFailures >= hc.failureThreshold {
		status.Healthy = false
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
