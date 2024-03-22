package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type apiConfig struct {
	fileserverHits int
}
type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

func main() {
	const filepathRoot = "."
	const port = "8080"
	const chirpDBPath = "chirps.json"
	db := NewDB(chirpDBPath)
	db.load()

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("/api/healthz", handlerReadiness)
	mux.HandleFunc("/admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("/api/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerValidateChirp)
	mux.HandleFunc("/api/chirp", apiCfg.handlerChirp(db))

	corsMux := middlewareCors(mux)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<html>
<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>
</html>`, cfg.fileserverHits)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}
func (cfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {

	var chirp Chirp
	err := json.NewDecoder(r.Body).Decode(&chirp)
	var profaneWords = []string{"kerfuffle",
		"sharbert",
		"fornax"}
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	if len(chirp.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp too long")
	}
	chirp.Body = replaceProfanity(chirp.Body, profaneWords)
	respondWithJSON(w, http.StatusOK, chirp)
}

func replaceProfanity(body string, profaneWords []string) string {
	for _, bodyWord := range strings.Fields(body) {
		for _, profaneWord := range profaneWords {
			if strings.ToLower(bodyWord) == profaneWord {
				body = strings.Replace(body, bodyWord, "****", -1)
			}
		}
	}
	return body
}

func (cfg *apiConfig) handlerChirp(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			idStr := r.URL.Query().Get("id")
			if idStr == "" {
				handlerGetAllChirps(w, r, db)
			} else {
				handlerGetChirp(w, r, db)
			}
		case http.MethodPost:
			handlerPostChirp(w, r, db)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
