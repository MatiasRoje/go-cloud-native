package main

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var store = make(map[string]string)

func main() {
	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/key/{key}", putHandler).Methods("PUT")
	api.HandleFunc("/key/{key}", getHandler).Methods("GET")
	api.HandleFunc("/key/{key}", deleteHandler).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", r))
}

// Handlers
// PUT {api-prefix}/key/{key}
func putHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := io.ReadAll(r.Body) // The request body has our value
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = PutKeyValue(key, string(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GET {api-prefix}/key/{key}
func getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := GetKeyValue(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

// DELETE {api-prefix}/key/{key}
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	err := DeleteKeyValue(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helpers
var ErrorNoSuchKey = errors.New("no such key")

func PutKeyValue(key, value string) error {
	store[key] = value

	return nil
}

func GetKeyValue(key string) (string, error) {
	value, ok := store[key]

	if !ok {
		return "", ErrorNoSuchKey
	}

	return value, nil
}

func DeleteKeyValue(key string) error {
	delete(store, key)

	return nil
}
