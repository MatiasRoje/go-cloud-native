package storage

import (
	"database/sql"
	"fmt"

	"github.com/MatiasRoje/go-cloud-native/internal/config"
)

type PostgresTransactionLogger struct {
	events chan<- Event // Write-only channel for sendings events
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

	exists, err := logger.verifyTableExists(logger.params.DBName)
	if err != nil {
		return nil, fmt.Errorf("failed to verify if table exists: %w", err)
	}

	if !exists {
		if err = logger.createTable(logger.params.DBName); err != nil {
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
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

func (l *PostgresTransactionLogger) verifyTableExists(tableName string) (bool, error) {
	var exists bool

	if err := l.db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = ?)", tableName).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to verify if table exists: %w", err)
	}

	return exists, nil
}

func (l *PostgresTransactionLogger) createTable(tableName string) error {
	if _, err := l.db.Exec("CREATE TABLE IF NOT EXISTS " + tableName + " (id SERIAL PRIMARY KEY, event_type VARCHAR(10), key VARCHAR(255), value VARCHAR(255))"); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

func (l *PostgresTransactionLogger) Run() {
	events := make(chan Event, 16)
	l.events = events

	errors := make(chan error, 1)
	l.errors = errors

	go func() {
		query := "INSERT INTO " + l.params.DBName + " (event_type, key, value) VALUES (?, ?, ?)"
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

		query := "SELECT event_type, key, value FROM " + l.params.DBName + " ORDER BY id"
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
