package app

import (
	"github.com/MatiasRoje/go-cloud-native/internal/models"
	"github.com/MatiasRoje/go-cloud-native/internal/storage"
)

type App struct {
	Logger storage.TransactionLogger
	Store  *models.LockableMap
}
