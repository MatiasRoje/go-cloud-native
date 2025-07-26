package main

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
