package main

import (
	"fmt"
	"net/http"
)

type Server struct {
	httpServer http.Server
}

func NewServer(port string, mux *http.ServeMux) Server {
	return Server{
		httpServer: http.Server{
			Addr:    ":" + port,
			Handler: mux,
		},
	}
}

func getHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func main() {
	const port = "8080"

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", getHealth)

	s := NewServer(port, mux)

	fmt.Printf("Serving on Port: %s\n", port)
	if err := s.httpServer.ListenAndServe(); err != nil {
		fmt.Errorf("Error Starting Server: %v", err)
	}
}
