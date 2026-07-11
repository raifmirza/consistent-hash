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

	//-------------------------------------------------
	// Registry
	//-------------------------------------------------

	registry := backend.NewRegistry()

	initialNodes := []*node.Node{
		{
			ID:      "server-8081",
			Address: "http://localhost:8081",
		},
		{
			ID:      "server-8082",
			Address: "http://localhost:8082",
		},
		{
			ID:      "server-8083",
			Address: "http://localhost:8083",
		},
	}

	for _, n := range initialNodes {

		if _, err := registry.Register(n); err != nil {
			log.Fatal(err)
		}
	}

	//-------------------------------------------------
	// Hash Ring
	//-------------------------------------------------

	ring := hash.NewHashRing(100)

	//-------------------------------------------------
	// Health Checker
	//-------------------------------------------------

	prober := health.NewHTTPProber(
		2 * time.Second,
	)

	checker := health.NewHealthChecker(
		registry,
		ring,
		5*time.Second,
		prober,
		3, // failures before removal
		2, // successes before recovery
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
		Addr:    ":8080",
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
