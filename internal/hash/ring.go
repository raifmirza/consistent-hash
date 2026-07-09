package hash

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/raifmirza/consistent-hash/internal/node"
)

var ErrEmptyRing = errors.New("hash ring is empty")

type HashRing struct {
	mu           sync.RWMutex
	replicas     int
	ring         map[uint32]string
	sortedHashes []uint32
	nodeHashes   map[string][]uint32
}

func NewHashRing(replicas int) *HashRing {
	return &HashRing{
		replicas: replicas,

		ring: make(map[uint32]string),

		sortedHashes: make([]uint32, 0),

		nodeHashes: make(map[string][]uint32),
	}
}

func (hr *HashRing) AddNode(n *node.Node) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if _, exists := hr.nodeHashes[n.ID]; exists {
		return
	}

	hashes := make([]uint32, 0, hr.replicas)

	for i := 0; i < hr.replicas; i++ {
		virtualID := fmt.Sprintf("%s#%d", n.ID, i)

		hash := Hash(virtualID)

		hr.ring[hash] = n.ID
		hr.sortedHashes = append(hr.sortedHashes, hash)
		hashes = append(hashes, hash)
	}

	hr.nodeHashes[n.ID] = hashes

	sort.Slice(hr.sortedHashes, func(i, j int) bool {
		return hr.sortedHashes[i] < hr.sortedHashes[j]
	})
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

func (hr *HashRing) GetBackendID(key string) (string, error) {

	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.sortedHashes) == 0 {
		return "", ErrEmptyRing
	}

	hash := Hash(key)

	idx := sort.Search(len(hr.sortedHashes), func(i int) bool {
		return hr.sortedHashes[i] >= hash
	})

	if idx == len(hr.sortedHashes) {
		idx = 0
	}

	return hr.ring[hr.sortedHashes[idx]], nil
}
