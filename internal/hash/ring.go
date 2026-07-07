package hash

import (
	"fmt"
	"sort"

	"github.com/raifmirza/consistent-hash/internal/node"
)

type HashRing struct {
	replicas   int
	ring       map[uint32]*node.Node
	sortedHash []uint32
}

func NewHashRing(replicas int) *HashRing {
	return &HashRing{
		replicas:   replicas,
		ring:       make(map[uint32]*node.Node),
		sortedHash: make([]uint32, 0),
	}
}

func (hr *HashRing) AddNode(n *node.Node) {
	for i := 0; i < hr.replicas; i++ {
		virtualID := fmt.Sprintf("%s#%d", n.ID, i)

		hash := Hash(virtualID)

		hr.ring[hash] = n
		hr.sortedHash = append(hr.sortedHash, hash)

	}
	sort.Slice(hr.sortedHash, func(i, j int) bool {
		return hr.sortedHash[i] < hr.sortedHash[j]
	})
}

func (hr *HashRing) GetNode(key string) (*node.Node, bool) {
	if len(hr.sortedHash) == 0 {
		return nil, false
	}
	hash := Hash(key)
	idx := sort.Search(len(hr.sortedHash), func(i int) bool {
		return hr.sortedHash[i] >= hash
	})

	if idx == len(hr.sortedHash) {
		idx = 0
	}
	nodeHash := hr.sortedHash[idx]
	return hr.ring[nodeHash], true
}

func (hr *HashRing) RemoveNode(n *node.Node) {

	for i := 0; i < hr.replicas; i++ {
		virtualID := fmt.Sprintf("%s#%d", n.ID, i)
		hash := Hash(virtualID)
		delete(hr.ring, hash)

		idx := sort.Search(len(hr.sortedHash), func(i int) bool {
			return hr.sortedHash[i] >= hash
		})

		if idx < len(hr.sortedHash) && hr.sortedHash[idx] == hash {
			hr.sortedHash = append(
				hr.sortedHash[:idx],
				hr.sortedHash[idx+1:]...,
			)
		}
	}
}

func (hr *HashRing) PrintRing() {
	for _, h := range hr.sortedHash {
		fmt.Printf("%10d -> %s\n", h, hr.ring[h].ID)
	}
}
