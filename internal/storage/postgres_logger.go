package storage

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/MatiasRoje/go-cloud-native/internal/config"
)

type PostgresTransactionLogger struct {
	events chan<- Event // Write-only channel for sending events
	errors <-chan error // Read-only channel for receiving errors
	db     *sql.DB
	params PostgresDBParams
}

type PostgresDBParams struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func NewPostgresTransactionLogger(cfg *config.Config) (*PostgresTransactionLogger, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("cannot open postgres transaction db: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("cannot ping postgres transaction db: %w", err)
	}

	logger := &PostgresTransactionLogger{
		db: db,
		params: PostgresDBParams{
			Host:     cfg.DBHost,
			Port:     cfg.DBPort,
			User:     cfg.DBUser,
			Password: cfg.DBPassword,
			DBName:   cfg.DBName,
		},
	}

	err = logger.initializeTransactionsTable()
	if err != nil {
		return nil, fmt.Errorf("failed to verify if table exists: %w", err)
	}

	return logger, nil
}

func (l *PostgresTransactionLogger) WritePut(key, value string) {
	l.events <- Event{
		EventType: EventPut,
		Key:       key,
		Value:     value,
	}
}

func (l *PostgresTransactionLogger) WriteDelete(key string) {
	l.events <- Event{
		EventType: EventDelete,
		Key:       key,
	}
}

func (l *PostgresTransactionLogger) Err() <-chan error {
	return l.errors
}

func (l *PostgresTransactionLogger) initializeTransactionsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			event_type VARCHAR(10),
			key VARCHAR(255),
			value VARCHAR(255) 
		)
	`

	if _, err := l.db.Exec(query); err != nil {
		return fmt.Errorf("error creating transactions table: %w", err)
	}

	log.Println("Transactions table successfully initialized")
	return nil
}

func (l *PostgresTransactionLogger) Run() {
	events := make(chan Event, 16)
	l.events = events

	errors := make(chan error, 1)
	l.errors = errors

	go func() {
		query := "INSERT INTO transactions (event_type, key, value) VALUES ($1, $2, $3)"
		for event := range events {
			_, err := l.db.Exec(query, event.EventType, event.Key, event.Value)
			if err != nil {
				errors <- err
			}
		}
	}()
}

func (l *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		defer close(outEvent)
		defer close(outError)

		query := "SELECT event_type, key, value FROM transactions ORDER BY id"
		rows, err := l.db.Query(query)
		if err != nil {
			outError <- fmt.Errorf("sql query error: %w", err)
			return
		}
		defer rows.Close()

		event := Event{}
		for rows.Next() {
			if err := rows.Scan(&event.EventType, &event.Key, &event.Value); err != nil {
				outError <- fmt.Errorf("error reading row: %w", err)
				return
			}
			outEvent <- event
		}

		if err := rows.Err(); err != nil {
			outError <- fmt.Errorf("transaction log read error: %w", err)
			return
		}
	}()

	return outEvent, outError
}

func (l *PostgresTransactionLogger) Close() error {
	if l.db != nil {
		return l.db.Close()
	}

	return nil
}
