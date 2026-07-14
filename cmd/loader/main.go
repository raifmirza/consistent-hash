package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/raifmirza/consistent-hash/internal/backend"
	"github.com/raifmirza/consistent-hash/internal/balancer"
	"github.com/raifmirza/consistent-hash/internal/config"
	"github.com/raifmirza/consistent-hash/internal/discovery"
	"github.com/raifmirza/consistent-hash/internal/hash"
	"github.com/raifmirza/consistent-hash/internal/health"
	"github.com/raifmirza/consistent-hash/internal/node"
)

func main() {

	// Graceful shutdown context
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	//-------------------------------------------------
	// Registry
	//-------------------------------------------------

	registry := backend.NewRegistry()

	for _, n := range cfg.Backends {

		if _, err := registry.Register(&node.Node{
			Address: n.Address,
			ID:      n.ID,
		}); err != nil {
			log.Fatal(err)
		}
	}

	//-------------------------------------------------
	// Hash Ring
	//-------------------------------------------------

	ring := hash.NewHashRing(cfg.HashRing.Replicas)

	//-------------------------------------------------
	// Health Checker
	//-------------------------------------------------

	prober := health.NewHTTPProber(
		cfg.Health.Timeout,
	)

	checker := health.NewHealthChecker(
		registry,
		ring,
		cfg.Health.Interval,
		prober,
		cfg.Health.FailureThreshold, // failures before removal
		cfg.Health.SuccessThreshold, // successes before recovery
	)

	go checker.Start(ctx)

	//-------------------------------------------------
	// Load Balancer
	//-------------------------------------------------

	lb := balancer.NewLoadBalancer(
		ring,
		registry,
		balancer.IpExtractor{},
	)

	//-------------------------------------------------
	// Discovery API
	//-------------------------------------------------

	discoveryHandler := discovery.NewHandler(
		registry,
		ring,
	)

	//-------------------------------------------------
	// HTTP Router
	//-------------------------------------------------

	mux := http.NewServeMux()

	// Load balancer
	mux.Handle("/", lb)

	// Discovery
	mux.HandleFunc(
		"POST /backends",
		discoveryHandler.Register,
	)

	mux.HandleFunc(
		"DELETE /backends/{id}",
		discoveryHandler.Delete,
	)

	server := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: mux,
	}

	//-------------------------------------------------
	// Start
	//-------------------------------------------------

	go func() {

		log.Println("Load Balancer listening on :8080")

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatal(err)
		}

	}()

	//-------------------------------------------------
	// Wait for shutdown
	//-------------------------------------------------

	<-ctx.Done()

	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server stopped")
}
