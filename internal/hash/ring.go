package hash

import (
	"fmt"
	"sort"
	"sync"

	"github.com/raifmirza/consistent-hash/internal/node"
)

type RingManager interface {
	AddNode(n *node.Node)
	RemoveNode(n *node.Node)
	GetNode(key string) (*node.Node, bool)
	PrintRing()
}

type HashRing struct {
	mu           sync.RWMutex
	replicas     int
	ring         map[uint32]*node.Node
	sortedHashes []uint32
	nodeHashes   map[string][]uint32
}

func NewHashRing(replicas int) *HashRing {
	return &HashRing{
		replicas:     replicas,
		ring:         make(map[uint32]*node.Node),
		sortedHashes: make([]uint32, 0),
		nodeHashes:   make(map[string][]uint32),
	}
}

func (hr *HashRing) AddNode(n *node.Node) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	// Prevent duplicate additions.
	if _, exists := hr.nodeHashes[n.ID]; exists {
		return
	}

	hashes := make([]uint32, 0, hr.replicas)

	for i := 0; i < hr.replicas; i++ {
		virtualID := fmt.Sprintf("%s#%d", n.ID, i)

		hash := Hash(virtualID)

		hr.ring[hash] = n
		hr.sortedHashes = append(hr.sortedHashes, hash)
		hashes = append(hashes, hash)
	}

	hr.nodeHashes[n.ID] = hashes

	sort.Slice(hr.sortedHashes, func(i, j int) bool {
		return hr.sortedHashes[i] < hr.sortedHashes[j]
	})
}

func (hr *HashRing) GetNode(key string) (*node.Node, bool) {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.sortedHashes) == 0 {
		return nil, false
	}

	hash := Hash(key)

	idx := sort.Search(len(hr.sortedHashes), func(i int) bool {
		return hr.sortedHashes[i] >= hash
	})

	if idx == len(hr.sortedHashes) {
		idx = 0
	}

	return hr.ring[hr.sortedHashes[idx]], true
}

func (hr *HashRing) RemoveNode(n *node.Node) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	hashes, exists := hr.nodeHashes[n.ID]
	if !exists {
		return
	}

	removeSet := make(map[uint32]struct{}, len(hashes))

	for _, hash := range hashes {
		delete(hr.ring, hash)
		removeSet[hash] = struct{}{}
	}

	newHashes := make([]uint32, 0, len(hr.sortedHashes)-len(hashes))

	for _, hash := range hr.sortedHashes {
		if _, shouldRemove := removeSet[hash]; !shouldRemove {
			newHashes = append(newHashes, hash)
		}
	}

	hr.sortedHashes = newHashes

	delete(hr.nodeHashes, n.ID)
}

func (hr *HashRing) PrintRing() {
	for _, h := range hr.sortedHashes {
		fmt.Printf("%10d -> %s\n", h, hr.ring[h].ID)
	}
}
