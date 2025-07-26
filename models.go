package main

import "sync"

type LockableMap struct {
	sync.RWMutex
	m map[string]string
}

var store = &LockableMap{
	m: make(map[string]string),
}
