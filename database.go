package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

type DB struct {
	Path string
	mux  sync.RWMutex
	Data map[int]Chirp
}

func NewDB(path string) *DB {
	return &DB{
		Path: path,
		Data: make(map[int]Chirp),
	}
}

func (db *DB) load() error {
	db.mux.Lock()
	defer db.mux.Unlock()

	file, err := os.ReadFile(db.Path)
	if err != nil {
		if os.IsNotExist(err) {
			_, err := os.Create(db.Path)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}

	if err := json.Unmarshal(file, &db.Data); err != nil {
		return err
	}

	return nil
}

func (db *DB) save() error {

	fmt.Printf(db.Path)

	file, err := json.Marshal(db.Data)
	if err != nil {
		fmt.Printf("Error marshalling data: %v\n", err)
		return err
	}
	err = os.WriteFile(db.Path, file, 0644)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			fmt.Println("Permission error")
		} else if errors.Is(err, os.ErrExist) {
			fmt.Println("File already exists error")
		} else if errors.Is(err, os.ErrClosed) {
			fmt.Println("File closed error")
		} else {
			fmt.Printf("Error writing file: %v\n", err)
		}
	}
	return err
}

func (db *DB) CreateChirp(chirp Chirp) error {

	db.mux.Lock()
	defer db.mux.Unlock()

	chirp.ID = len(db.Data) + 1
	db.Data[chirp.ID] = chirp

	err := db.save()
	if err != nil {
		fmt.Printf("Error saving data: %v\n", err)
	}
	return err
}

func (db *DB) DeleteChirp(id int) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	if _, exists := db.Data[id]; !exists {
		return fmt.Errorf("Chirp with ID %d does not exist", id)
	}
	delete(db.Data, id)
	return db.save()
}
func (db *DB) GetAllChirps() []Chirp {
	db.mux.RLock()
	defer db.mux.RUnlock()

	chirps := make([]Chirp, 0, len(db.Data))
	for _, chirp := range db.Data {
		chirps = append(chirps, chirp)
	}
	return chirps
}

func (db *DB) GetChirp(id int) (Chirp, bool) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	chirp, exists := db.Data[id]
	fmt.Println(db.Data)
	return chirp, exists
}

func (db *DB) UpdateChirp(chirp Chirp) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	if _, exists := db.Data[chirp.ID]; !exists {
		return fmt.Errorf("Chirp with ID %d does not exist", chirp.ID)
	}
	db.Data[chirp.ID] = chirp
	return db.save()
}
