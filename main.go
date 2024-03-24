package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type apiConfig struct {
	fileserverHits int
	jwtSecret      string
	polkaSecret    string
}
type Chirp struct {
	ID     int    `json:"id"`
	Body   string `json:"body"`
	Author string `json:"author"`
}
type User struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Email             string `json:"email"`
	Password          string `json:"password"`
	IS_CHIRPY_PREMIUM bool   `json:"is_chirpy_premium"`
}
type Webhook struct {
	Event string `json:"event"`
	Data  struct {
		UserID int `json:"user_id"`
	} `json:"data"`
}

func main() {
	const filepathRoot = "."
	const port = "8080"
	const chirpDBPath = "chirps.json"
	chirpDB := NewDB(chirpDBPath)
	chirpDB.load()

	apiCfg := apiConfig{
		fileserverHits: 0,
		jwtSecret:      os.Getenv("JWT_SECRET"),
		polkaSecret:    os.Getenv("POLKA_SECRET"),
	}

	mux := http.NewServeMux()
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("/api/healthz", handlerReadiness)
	mux.HandleFunc("/admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("/api/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerValidateChirp)
	mux.Handle("/api/chirp", apiCfg.authMiddleware(apiCfg.handlerChirp(chirpDB)))
	mux.Handle("/api/chirp/", apiCfg.authMiddleware(apiCfg.handlerChirp(chirpDB)))
	mux.Handle("/api/users", apiCfg.authMiddleware(apiCfg.handlerUsers(chirpDB)))
	mux.Handle("/api/users/", apiCfg.authMiddleware(apiCfg.handlerUsers(chirpDB)))
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin(chirpDB))
	mux.HandleFunc("POST /api/register", apiCfg.handlerRegister(chirpDB))
	mux.HandleFunc("/api/polka/webhooks", apiCfg.handlerPolkaWebhooks(chirpDB))
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
func (cfg *apiConfig) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		authHeader := r.Header.Get("Authorization")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.jwtSecret), nil
		})
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
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
		case http.MethodDelete:
			segments := strings.Split(r.URL.Path, "/")
			if len(segments) < 4 {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}
			id := segments[3]
			cfg.handlerDeleteChirp(w, r, chirpDB, id)
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
func (cfg *apiConfig) handlerRegister(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		_, exists := db.GetUserByEmail(user.Email)
		if exists {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}
		user.Password = string(hashedPassword)
		db.CreateUser(user)
		respondWithJSON(w, http.StatusCreated, user)
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
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"issuer":    "chirpy",
			"email":     user.Email,
			"issuedAt":  time.Now().Unix(),
			"expiresAt": time.Now().Add(time.Hour * 24).Unix(),
			"subject":   strconv.Itoa(storedUser.ID),
		})
		tokenString, err := token.SignedString([]byte(cfg.jwtSecret))
		if err != nil {
			http.Error(w, "Error signing token", http.StatusInternalServerError)
			return
		}
		responsPayload := map[string]string{"id": strconv.Itoa(storedUser.ID), "token": tokenString, "email": storedUser.Email}
		respondWithJSON(w, http.StatusOK, responsPayload)
	}

}

func (cfg *apiConfig) handlerPolkaWebhooks(db *DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]
		if tokenString != cfg.polkaSecret {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var webhook Webhook
		err := json.NewDecoder(r.Body).Decode(&webhook)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		if webhook.Event == "user.upgraded" {
			user, exists := db.GetUserByID(webhook.Data.UserID)
			if !exists {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			user.IS_CHIRPY_PREMIUM = true
			db.UpdateUser(user)
		}
		respondWithJSON(w, http.StatusOK, webhook)
	}
}
