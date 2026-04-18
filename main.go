package main

import (
	"fmt"
	"net/http"
)

type Server struct {
	httpServer http.Server
}

func NewServer(port string, serveMux *http.ServeMux) Server {
	return Server{
		httpServer: http.Server{
			Addr:    ":" + port,
			Handler: serveMux,
		},
	}
}

func main() {
	port := "8080"
	mux := http.NewServeMux()
	s := NewServer(port, mux)

	mux.Handle("/", http.FileServer(http.Dir(".")))

	fmt.Printf("Serving on Port: %s\n", port)
	if err := s.httpServer.ListenAndServe(); err != nil {
		fmt.Errorf("Error Starting Server: %v", err)
	}
}
