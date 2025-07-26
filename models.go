package main

import "sync"

const (
	MaxKeySize   = 256
	MaxValueSize = 4096
)

type LockableMap struct {
	sync.RWMutex
	m map[string]string
}

var store = &LockableMap{
	m: make(map[string]string),
}
