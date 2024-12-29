package database

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	ConnectionString string
}

func New(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}
