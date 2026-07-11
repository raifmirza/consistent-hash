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

func (r *Registry) Register(n *node.Node) (*Backend, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.backends[n.ID]; exists {
		return nil, fmt.Errorf("backend %q already exists", n.ID)
	}

	target, err := url.Parse(n.Address)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	backend := &Backend{
		Node:  n,
		Proxy: proxy,
	}

	r.backends[n.ID] = backend

	return backend, nil
}

func (r *Registry) Lookup(id string) (*Backend, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	backend, ok := r.backends[id]
	if !ok {
		return nil, fmt.Errorf("backend %q not found", id)
	}

	return backend, nil
}

func (r *Registry) Remove(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.backends[id]; !ok {
		return fmt.Errorf("backend %q not found", id)
	}

	delete(r.backends, id)

	return nil
}

func (r *Registry) Exists(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.backends[id]
	return ok
}

func (r *Registry) List() []*Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()

	backends := make([]*Backend, 0, len(r.backends))

	for _, backend := range r.backends {
		backends = append(backends, backend)
	}

	return backends
}
