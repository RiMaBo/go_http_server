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

func main() {
	const port = "8080"

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))

	s := NewServer(port, mux)

	fmt.Printf("Serving on Port: %s\n", port)
	if err := s.httpServer.ListenAndServe(); err != nil {
		fmt.Errorf("Error Starting Server: %v", err)
	}
}
