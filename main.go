package main

import _ "github.com/lib/pq"
import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/RiMaBo/go_http_server/internal/database"
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
	err := godotenv.Load()
	if err != nil {
		fmt.Errorf("Error Loading .env File")
	}

	dbURL := os.Getenv("DB_URL")
	if len(dbURL) < 1 {
		fmt.Errorf("DB_URL Must Be Set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Errorf("Error Opening Database: %v", err)
	}
	defer db.Close()

	dbQueries := database.New(db)

	const port = "8080"
	apiCfg := &apiConfig{
		db: dbQueries,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", getHealth)
	mux.HandleFunc("GET /admin/metrics", apiCfg.getMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	s := NewServer(port, mux)

	fmt.Printf("Serving on Port: %s\n", port)
	if err := s.httpServer.ListenAndServe(); err != nil {
		fmt.Errorf("Error Starting Server: %v", err)
	}
}
