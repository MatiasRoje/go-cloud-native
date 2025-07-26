package storage

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

type FileTransactionLogger struct {
	events       chan<- Event // Write-only channel for sendings events
	errors       <-chan error // Read-only channel for receiving errors
	lastSequence uint64
	file         *os.File
}

func NewFileTransactionLogger(filename string) (*FileTransactionLogger, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaction log file: %w", err)
	}

	return &FileTransactionLogger{
		file: file,
	}, nil
}

func (l *FileTransactionLogger) WritePut(key, value string) {
	encodedKey := base64.StdEncoding.EncodeToString([]byte(key))
	encodedValue := base64.StdEncoding.EncodeToString([]byte(value))

	l.events <- Event{
		EventType: EventPut,
		Key:       encodedKey,
		Value:     encodedValue,
	}
}

func (l *FileTransactionLogger) WriteDelete(key string) {
	encodedKey := base64.StdEncoding.EncodeToString([]byte(key))

	l.events <- Event{
		EventType: EventDelete,
		Key:       encodedKey,
	}
}

func (l *FileTransactionLogger) Err() <-chan error {
	return l.errors
}

func (l *FileTransactionLogger) Run() {
	events := make(chan Event, 16)
	l.events = events

	errors := make(chan error, 1)
	l.errors = errors

	go func() {
		for event := range events {

			l.lastSequence++
			event.Sequence = l.lastSequence

			jsonLine, err := json.Marshal(event)
			if err != nil {
				errors <- fmt.Errorf("json marshal error: %w", err)
				return
			}

			_, err = fmt.Fprintf(l.file, "%s\n", jsonLine)
			if err != nil {
				errors <- err
				return
			}
		}
	}()
}

func (l *FileTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	scanner := bufio.NewScanner(l.file)
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		defer close(outEvent)
		defer close(outError)
		for scanner.Scan() {
			var event Event
			line := scanner.Text()

			if err := json.Unmarshal([]byte(line), &event); err != nil {
				outError <- fmt.Errorf("json unmarshal error: %w", err)
				return
			}

			decodedKey, err := base64.StdEncoding.DecodeString(event.Key)
			if err != nil {
				outError <- fmt.Errorf("key decode error: %w", err)
				return
			}
			event.Key = string(decodedKey)

			if event.EventType == EventPut {
				decodedValue, err := base64.StdEncoding.DecodeString(event.Value)
				if err != nil {
					outError <- fmt.Errorf("value decode error: %w", err)
					return
				}
				event.Value = string(decodedValue)
			}

			// Sanity check! Are the sequence numbers in increasing order?
			if l.lastSequence >= event.Sequence {
				outError <- fmt.Errorf("transaction numbers out of sequence")
				return
			}

			l.lastSequence = event.Sequence

			outEvent <- event
		}

		if err := scanner.Err(); err != nil {
			outError <- fmt.Errorf("transaction log read error: %w", err)
			return
		}
	}()

	return outEvent, outError
}

func (l *FileTransactionLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
