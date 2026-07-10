package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/raifmirza/consistent-hash/internal/backend"
	"github.com/raifmirza/consistent-hash/internal/balancer"
	"github.com/raifmirza/consistent-hash/internal/hash"
	"github.com/raifmirza/consistent-hash/internal/health"
	"github.com/raifmirza/consistent-hash/internal/node"
)

func main() {
	nodes := []*node.Node{
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
	registry := backend.NewRegistry()

	for _, node := range nodes {
		_, err := registry.Register(node)
		if err != nil {
			log.Fatal(err)
		}
	}

	ring := hash.NewHashRing(100)
	prober := health.NewHttpProber(time.Second)
	checker := health.NewHealthChecker(ring, 5*time.Second, prober, nodes, 3, 2)

	ctx := context.Background()
	go checker.Start(ctx)

	lb := balancer.NewLoadBalancer(ring, registry, balancer.IpExtractor{})

	mux := http.NewServeMux()
	mux.Handle("/", lb)
	srv := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		log.Println("Load balancer listening on :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen error: %v", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	log.Println("server exited gracefully")
}
