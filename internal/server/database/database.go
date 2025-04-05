// Package database contains postgresql db factory
package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	ConnectionString string
	RetryAttempts    []time.Duration
}

type PgxDatabaseFactory struct {
	cfg Config
}

func NewPgxDatabaseFactory(cfg Config) *PgxDatabaseFactory {
	return &PgxDatabaseFactory{
		cfg: cfg,
	}
}

func (f *PgxDatabaseFactory) Create() (*sql.DB, error) {
	db, err := sql.Open("pgx", f.cfg.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}
