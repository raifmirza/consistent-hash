package discovery

import (
	"encoding/json"
	"net/http"

	"github.com/raifmirza/consistent-hash/internal/backend"
	"github.com/raifmirza/consistent-hash/internal/hash"
	"github.com/raifmirza/consistent-hash/internal/node"
)

type Handler struct {
	registry *backend.Registry
	ring     hash.RingManager
}

func NewHandler(registry *backend.Registry, ring hash.RingManager) *Handler {
	return &Handler{
		registry: registry,
		ring:     ring,
	}
}

type RegisterRequest struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	n := &node.Node{
		ID:      req.ID,
		Address: req.Address,
	}

	if _, err := h.registry.Register(n); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	backend, err := h.registry.Lookup(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	h.ring.RemoveNode(backend.Node)
	if err := h.registry.Remove(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
