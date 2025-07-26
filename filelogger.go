package main

import (
	"bufio"
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
	l.events <- Event{
		EventType: EventPut,
		Key:       key,
		Value:     value,
	}
}

func (l *FileTransactionLogger) WriteDelete(key string) {
	l.events <- Event{
		EventType: EventDelete,
		Key:       key,
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

			_, err := fmt.Fprintf(l.file, "%d\t%d\t%s\t%s\n", l.lastSequence, event.EventType, event.Key, event.Value)
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
		var event Event
		defer close(outEvent)
		defer close(outError)
		for scanner.Scan() {
			line := scanner.Text()

			if _, err := fmt.Sscanf(line, "%d\t%d\t%s\t%s", &event.Sequence, &event.EventType, &event.Key, &event.Value); err != nil {
				outError <- fmt.Errorf("input parse error: %w", err)
				return
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
