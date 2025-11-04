package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/NHMosko/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	platform := os.Getenv("PLATFORM")
	dbURL := os.Getenv("DB_URL")
	db,err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	const prefix = "/app/"
	const filepathRoot = "."
	const port = "8080"
	apiCfg := apiConfig{
		dbQueries: dbQueries,
		platform: platform,
	}

	mux := http.NewServeMux()
	mux.Handle(prefix, apiCfg.middleMetricsInc(handle(prefix, filepathRoot)))
	mux.HandleFunc("GET /admin/metrics", apiCfg.getMetrics)
	mux.HandleFunc("GET /api/chirps", apiCfg.getChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirpByID)
	mux.HandleFunc("GET /api/healthz", getHealth)
	mux.HandleFunc("POST /admin/reset", apiCfg.handleReset)
	mux.HandleFunc("POST /api/users", apiCfg.createUser)
	mux.HandleFunc("POST /api/chirps", apiCfg.createChirp)
	mux.HandleFunc("POST /api/login", apiCfg.login)

	server := http.Server{
		Handler: mux,
		Addr: ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func handle(prefix string, filepathRoot string) http.Handler {
	return http.StripPrefix(prefix, http.FileServer(http.Dir(filepathRoot)))
}

func getHealth(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
}


