package main

import _ "github.com/lib/pq"
import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"

	"github.com/RiMaBo/go_http_server/internal/database"
)

type Server struct {
	httpServer http.Server
}

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
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
		fmt.Errorf("DB_URL must be set")
	}

	platform := os.Getenv("PLATFORM")
	if len(platform) < 1 {
		fmt.Errorf("PLATFORM must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Errorf("Error opening database: %v", err)
	}
	defer db.Close()

	dbQueries := database.New(db)

	const port = "8080"
	apiCfg := &apiConfig{
		db:       dbQueries,
		platform: platform,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET  /admin/metrics", apiCfg.getMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)
	mux.HandleFunc("GET  /api/healthz", getHealth)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUsers)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)

	s := NewServer(port, mux)

	fmt.Printf("Serving on port: %s\n", port)
	if err := s.httpServer.ListenAndServe(); err != nil {
		fmt.Errorf("Error starting server: %v", err)
	}
}
