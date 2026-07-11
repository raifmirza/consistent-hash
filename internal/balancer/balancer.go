package balancer

import (
	"net/http"

	"github.com/raifmirza/consistent-hash/internal/backend"
	"github.com/raifmirza/consistent-hash/internal/hash"
)

type BackendRegistry interface {
	Lookup(id string) (*backend.Backend, error)
}

type Router interface {
	GetBackendID(key string) (string, error)
}

type LoadBalancer struct {
	router hash.Router

	registry BackendRegistry

	extractor KeyExtractor
}

func NewLoadBalancer(
	router hash.Router,
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

	backend, err := lb.registry.Lookup(backendID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	backend.Proxy.ServeHTTP(w, r)
}
