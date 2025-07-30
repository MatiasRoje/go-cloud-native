package storage

import (
	"fmt"

	"github.com/MatiasRoje/go-cloud-native/internal/config"
	"github.com/MatiasRoje/go-cloud-native/internal/models"
	_ "github.com/lib/pq"
)

type EventType byte

type Event struct {
	Sequence  uint64
	EventType EventType
	Key       string
	Value     string
}

const (
	_                     = iota
	EventDelete EventType = iota
	EventPut
)

type TransactionLogger interface {
	WritePut(key, value string)
	WriteDelete(key string)
	Err() <-chan error

	ReadEvents() (<-chan Event, <-chan error)

	Run()

	Close() error
}

func InitLogger(path, loggerType string, cfg *config.Config) (TransactionLogger, error) {
	var logger TransactionLogger
	var err error
	switch loggerType {
	case "file":
		logger, err = NewFileTransactionLogger(path)
		if err != nil {
			return nil, fmt.Errorf("cannot initialize transaction logger: %w", err)
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
					err = models.DeleteKeyValue(event.Key)
				case EventPut:
					err = models.PutKeyValue(event.Key, event.Value)
				}
			}
		}

		logger.Run()

		return logger, nil
	case "postgres":

		logger, err = NewPostgresTransactionLogger(cfg)
		if err != nil {
			return nil, fmt.Errorf("cannon initialize postgres transaction logger: %w", err)
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
					err = models.DeleteKeyValue(event.Key)
				case EventPut:
					err = models.PutKeyValue(event.Key, event.Value)
				}
			}
		}

		logger.Run()

		return logger, nil
	default:
		return nil, fmt.Errorf("unknown logger type: %s", loggerType)
	}
}
