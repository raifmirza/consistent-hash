package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sync/atomic"
)

type Response struct {
	ServerID string `json:"server_id"`
	Path     string `json:"path"`
	Method   string `json:"method"`
	Requests uint64 `json:"requests"`
}

func main() {

	port := flag.String("port", "8081", "server port")
	flag.Parse()

	serverID := "server-" + *port

	var requestCount atomic.Uint64

	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Demo endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		count := requestCount.Add(1)

		resp := Response{
			ServerID: serverID,
			Path:     r.URL.Path,
			Method:   r.Method,
			Requests: count,
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Printf("%s listening on :%s", serverID, *port)

	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		log.Fatal(err)
	}
}
