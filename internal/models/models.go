package models

import (
	"errors"
	"fmt"
	"sync"
)

const (
	MaxKeySize   = 256
	MaxValueSize = 4096
)

type LockableMap struct {
	sync.RWMutex
	M map[string]string
}

var Store = &LockableMap{
	M: make(map[string]string),
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
	Store.Lock()
	defer Store.Unlock()

	Store.M[key] = value

	return nil
}

func GetKeyValue(key string) (string, error) {
	Store.RLock()
	defer Store.RUnlock()

	value, ok := Store.M[key]
	if !ok {
		return "", ErrorNoSuchKey
	}

	return value, nil
}

func DeleteKeyValue(key string) error {
	if len(key) > MaxKeySize {
		return fmt.Errorf("key exceeds maximum size of %d bytes", MaxKeySize)
	}
	Store.Lock()
	defer Store.Unlock()

	delete(Store.M, key)

	return nil
}
