package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type apiConfig struct {
	fileserverHits int
}
type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	const filepathRoot = "."
	const port = "8080"
	const chirpDBPath = "chirps.json"
	chirpDB := NewDB(chirpDBPath)
	chirpDB.load()

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("/api/healthz", handlerReadiness)
	mux.HandleFunc("/admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("/api/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerValidateChirp)
	mux.HandleFunc("/api/chirp", apiCfg.handlerChirp(chirpDB))
	mux.HandleFunc("/api/chirp/", apiCfg.handlerChirp(chirpDB))
	mux.HandleFunc("/api/users", apiCfg.handlerUsers(chirpDB))
	mux.HandleFunc("/api/users/", apiCfg.handlerUsers(chirpDB))
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin(chirpDB))
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

func (cfg *apiConfig) handlerChirp(chirpDB *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			segments := strings.Split(r.URL.Path, "/")
			if len(segments) > 3 {
				id := segments[3]
				handlerGetChirp(w, r, chirpDB, id)
			} else {
				handlerGetAllChirps(w, r, chirpDB)
			}
		case http.MethodPost:
			handlerPostChirp(w, r, chirpDB)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (cfg *apiConfig) handlerUsers(chirpDB *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlerGetUsers(w, r, chirpDB)
		case http.MethodPost:
			handlerPostUser(w, r, chirpDB)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
func (cfg *apiConfig) handlerLogin(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		storedUser, exists := db.GetUserByEmail(user.Email)
		if !exists {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
		if err != nil {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}
		fmt.Println("User logged in")
	}
}
