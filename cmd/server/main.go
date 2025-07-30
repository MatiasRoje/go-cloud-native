package main

import (
	"log"
	"net/http"

	"github.com/MatiasRoje/go-cloud-native/internal/api"
	"github.com/MatiasRoje/go-cloud-native/internal/app"
	"github.com/MatiasRoje/go-cloud-native/internal/config"
	"github.com/MatiasRoje/go-cloud-native/internal/models"
	"github.com/MatiasRoje/go-cloud-native/internal/storage"
	"github.com/gorilla/mux"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := storage.InitLogger("transaction.log", "postgres", config)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Close()

	// NOTE: We could move some of this logic to the app package and do something like app.Run() here
	app := &app.App{
		Logger: logger,
		Store:  models.Store,
		Config: config,
	}

	r := mux.NewRouter()
	apiRouter := r.PathPrefix("/api/v1").Subrouter()

	kvStoreHandler := api.NewKVStoreHandler(app)
	apiRouter.HandleFunc("/key/{key}", kvStoreHandler.PutHandler).Methods("PUT")
	apiRouter.HandleFunc("/key/{key}", kvStoreHandler.GetHandler).Methods("GET")
	apiRouter.HandleFunc("/key/{key}", kvStoreHandler.DeleteHandler).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", r))
}
