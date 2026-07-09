package backend

import (
	"fmt"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/raifmirza/consistent-hash/internal/node"
)

type Registry struct {
	mu sync.RWMutex

	backends map[string]*Backend
}

func NewRegistry() *Registry {
	return &Registry{
		backends: make(map[string]*Backend),
	}
}

func (registry *Registry) Register(n *node.Node) (*Backend, error) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	if _, exists := registry.backends[n.ID]; exists {
		return nil, fmt.Errorf("backend %s already exists", n.ID)
	}
	target, err := url.Parse(n.Address)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	backend := &Backend{Node: n, Proxy: proxy}
	registry.backends[n.ID] = backend

	return backend, nil
}

func (registry *Registry) Remove(id string) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	delete(registry.backends, id)
}

func (registry *Registry) GetBackend(id string) (*Backend, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	backend, ok := registry.backends[id]
	return backend, ok
}
