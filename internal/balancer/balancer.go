package balancer

import (
	"net/http"

	b "github.com/raifmirza/consistent-hash/internal/backend"
)

type BackendRegistry interface {
	GetBackend(id string) (*b.Backend, bool)
}

type Router interface {
	GetBackendID(key string) (string, error)
}

type LoadBalancer struct {
	router Router

	registry BackendRegistry

	extractor KeyExtractor
}

func NewLoadBalancer(
	router Router,
	registry BackendRegistry,
	extractor KeyExtractor,
) *LoadBalancer {

	return &LoadBalancer{
		router:    router,
		registry:  registry,
		extractor: extractor,
	}
}

func (lb *LoadBalancer) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {

	key, err := lb.extractor.Extract(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	backendID, err := lb.router.GetBackendID(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	backend, ok := lb.registry.GetBackend(backendID)
	if !ok {
		http.Error(w, "backend not found", http.StatusServiceUnavailable)
		return
	}

	backend.Proxy.ServeHTTP(w, r)
}
