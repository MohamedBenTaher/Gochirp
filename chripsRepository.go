package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

type DB struct {
	Path      string
	mux       sync.RWMutex
	ChirpData map[int]Chirp
	UserData  map[int]User
}

func NewDB(path string) *DB {
	return &DB{
		Path:      path,
		ChirpData: make(map[int]Chirp),
		UserData:  make(map[int]User),
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

	if err := json.Unmarshal(file, &db.ChirpData); err != nil {
		return err
	}

	return nil
}

func (db *DB) save(dataType string) error {
	var file []byte
	var err error

	if dataType == "users" {
		file, err = json.Marshal(db.UserData)
		if err != nil {
			return err
		}
	} else {
		file, err = json.Marshal(db.ChirpData)
		if err != nil {
			return err
		}
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

	chirp.ID = len(db.ChirpData) + 1
	db.ChirpData[chirp.ID] = chirp

	err := db.save("chirps")
	if err != nil {
		fmt.Printf("Error saving data: %v\n", err)
	}
	return err
}

func (db *DB) DeleteChirp(id int) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	if _, exists := db.ChirpData[id]; !exists {
		return fmt.Errorf("Chirp with ID %d does not exist", id)
	}
	delete(db.ChirpData, id)
	return db.save("chirps")
}
func (db *DB) GetAllChirps() []Chirp {
	db.mux.RLock()
	defer db.mux.RUnlock()

	chirps := make([]Chirp, 0, len(db.ChirpData))
	for _, chirp := range db.ChirpData {
		chirps = append(chirps, chirp)
	}
	return chirps
}

func (db *DB) GetChirp(id int) (Chirp, bool) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	chirp, exists := db.ChirpData[id]
	fmt.Println(db.ChirpData)
	return chirp, exists
}

func (db *DB) UpdateChirp(chirp Chirp) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	if _, exists := db.ChirpData[chirp.ID]; !exists {
		return fmt.Errorf("Chirp with ID %d does not exist", chirp.ID)
	}
	db.ChirpData[chirp.ID] = chirp
	return db.save("chirps")
}

func (db *DB) CreateUser(user User) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	for _, existingUser := range db.UserData {
		if existingUser.Email == user.Email {
			return fmt.Errorf("User with email %s already exists", user.Email)
		}
	}
	user.ID = len(db.UserData) + 1
	db.UserData[user.ID] = user

	err := db.save("users")
	if err != nil {
		fmt.Printf("Error saving data: %v\n", err)
	}
	return err
}
func (db *DB) GetAllUsers() []User {
	db.mux.RLock()
	defer db.mux.RUnlock()

	users := make([]User, 0, len(db.UserData))
	for _, user := range db.UserData {
		users = append(users, user)
	}
	return users
}
func (db *DB) GetUserByEmail(email string) (User, bool) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	for _, user := range db.UserData {
		if user.Email == email {
			return user, true
		}
	}
	return User{}, false
}
