package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
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

	log.Printf("%s listening on :%s", serverID, *port)

	log.Fatal(http.ListenAndServe(":"+*port, mux))
}
