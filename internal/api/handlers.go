package api

import (
	"io"
	"net/http"

	"github.com/MatiasRoje/go-cloud-native/internal/app"
	"github.com/MatiasRoje/go-cloud-native/internal/models"
	"github.com/gorilla/mux"
)

type KVStoreHandler struct {
	App *app.App
}

func NewKVStoreHandler(app *app.App) *KVStoreHandler {
	return &KVStoreHandler{
		App: app,
	}
}

// PUT {api-prefix}/key/{key}
func (h *KVStoreHandler) PutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := io.ReadAll(r.Body) // The request body has our value
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = models.PutKeyValue(key, string(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.App.Logger.WritePut(key, string(value))

	w.WriteHeader(http.StatusCreated)
}

// GET {api-prefix}/key/{key}
func (h *KVStoreHandler) GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := models.GetKeyValue(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

// DELETE {api-prefix}/key/{key}
func (h *KVStoreHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	err := models.DeleteKeyValue(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.App.Logger.WriteDelete(key)

	w.WriteHeader(http.StatusNoContent)
}
