package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
func handlerGetChirp(w http.ResponseWriter, r *http.Request, db *DB) {
	fmt.Println("in handlerGetChirp")
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
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
