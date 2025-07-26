package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

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

	logger.WritePut(key, string(value))

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

	logger.WriteDelete(key)

	w.WriteHeader(http.StatusNoContent)
}

// Helpers
var ErrorNoSuchKey = errors.New("no such key")

func PutKeyValue(key, value string) error {
	if len(key) > MaxKeySize {
		return fmt.Errorf("key exceeds maximum size of %d bytes", MaxKeySize)
	}
	if len(value) > MaxValueSize {
		return fmt.Errorf("value exceeds maximum size of %d bytes", MaxValueSize)
	}
	store.Lock()
	defer store.Unlock()

	store.m[key] = value

	return nil
}

func GetKeyValue(key string) (string, error) {
	store.RLock()
	defer store.RUnlock()

	value, ok := store.m[key]
	if !ok {
		return "", ErrorNoSuchKey
	}

	return value, nil
}

func DeleteKeyValue(key string) error {
	if len(key) > MaxKeySize {
		return fmt.Errorf("key exceeds maximum size of %d bytes", MaxKeySize)
	}
	store.Lock()
	defer store.Unlock()

	delete(store.m, key)

	return nil
}
