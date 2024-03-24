package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
func handlerPostChirp(w http.ResponseWriter, r *http.Request, db *DB) {
	fmt.Println("in handlerPostChirp")
	var chirp Chirp
	err := json.NewDecoder(r.Body).Decode(&chirp)
	if err != nil {
		fmt.Printf("Error decoding chirp: %v\n", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := db.CreateChirp(chirp); err != nil {
		fmt.Println("Error creating chirp")
		http.Error(w, "Failed to create chirp", http.StatusInternalServerError)
		return
	}
	fmt.Println("Chirp created")

	respondWithJSON(w, http.StatusCreated, chirp)
}
func handlerGetChirp(w http.ResponseWriter, r *http.Request, db *DB, ChirpID string) {

	id, err := strconv.Atoi(ChirpID)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	fmt.Printf("ID: %d\n", id)

	chirp, exists := db.GetChirp(id)
	if !exists {
		http.Error(w, "Chirp not found", http.StatusNotFound)
		return
	}
	fmt.Println("Chirp found")

	respondWithJSON(w, http.StatusOK, chirp)
}
func handlerGetAllChirps(w http.ResponseWriter, r *http.Request, db *DB) {
	chirps := db.GetAllChirps()
	respondWithJSON(w, http.StatusOK, chirps)
}

func handlerDeleteChirp(w http.ResponseWriter, r *http.Request, db *DB, ChirpID string) {
	id, err := strconv.Atoi(ChirpID)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	fmt.Printf("ID: %d\n", id)

	if err := db.DeleteChirp(id); err != nil {
		http.Error(w, "Failed to delete chirp", http.StatusInternalServerError)
		return
	}
	fmt.Println("Chirp deleted")

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
func handlerUpdateChirp(w http.ResponseWriter, r *http.Request, db *DB) {
	var chirp Chirp
	err := json.NewDecoder(r.Body).Decode(&chirp)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := db.UpdateChirp(chirp); err != nil {
		http.Error(w, "Failed to update chirp", http.StatusInternalServerError)
		return
	}
	fmt.Println("Chirp updated")

	respondWithJSON(w, http.StatusOK, chirp)
}

func handlerGetUsers(w http.ResponseWriter, r *http.Request, db *DB) {
	users := db.GetAllUsers()
	respondWithJSON(w, http.StatusOK, users)
}
func handlerPostUser(w http.ResponseWriter, r *http.Request, db *DB) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hash)

	if err := db.CreateUser(user); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}
	fmt.Println("User created")

	respondWithJSON(w, http.StatusCreated, user)
}
