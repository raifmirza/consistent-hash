package hash

import "github.com/raifmirza/consistent-hash/internal/node"

type Router interface {
	GetBackendID(key string) (string, error)
}

type RingManager interface {
	AddNode(node *node.Node)
	RemoveNode(node *node.Node)
}
