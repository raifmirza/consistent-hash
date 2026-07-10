package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Response struct {
	Server string `json:"server"`
	Path   string `json:"path"`
}

func main() {

	port := flag.String("port", "8081", "")
	flag.Parse()

	serverID := "server-" + *port

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		resp := Response{
			Server: serverID,
			Path:   r.URL.Path,
		}

		json.NewEncoder(w).Encode(resp)
	})

	srv := http.Server{Addr: ":" + *port, Handler: mux}

	go func() {
		log.Printf("%s listening on :%s", serverID, *port)
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
