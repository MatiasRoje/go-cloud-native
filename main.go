package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var logger TransactionLogger

func main() {
	err := initializeTransactionLog()
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/key/{key}", putHandler).Methods("PUT")
	api.HandleFunc("/key/{key}", getHandler).Methods("GET")
	api.HandleFunc("/key/{key}", deleteHandler).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", r))
}

func initializeTransactionLog() error {
	var err error
	logger, err = NewFileTransactionLogger("transaction.log")
	if err != nil {
		return fmt.Errorf("cannot initialize transaction logger: %w", err)
	}

	events, errors := logger.ReadEvents()

	event := Event{}
	ok := true

	for ok && err == nil {
		select {
		case err, ok = <-errors:
		case event, ok = <-events:
			switch event.EventType {
			case EventDelete:
				err = DeleteKeyValue(event.Key)
			case EventPut:
				err = PutKeyValue(event.Key, event.Value)
			}
		}
	}

	logger.Run()

	return nil
}
