package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type contextKey int

const (
	transactionKey contextKey = iota
)

const (
	setupDatabaseRequest = `
		create table if not exists metrics
		(
			key           varchar(63) not null primary key,
			gauge_value   double precision,
			counter_value bigint
			check ((counter_value is null) != (gauge_value is null))
		);`
)

type DBFactory interface {
	Create() (*sql.DB, error)
}

type DBStorage struct {
	db            *sql.DB
	logger        *zap.Logger
	retryAttempts []time.Duration
}

func New(dbFactory DBFactory, retryAttempts []time.Duration, logger *zap.Logger) (*DBStorage, error) {
	db, err := dbFactory.Create()
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}
	_, err = db.Exec(setupDatabaseRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}
	return &DBStorage{
		db:            db,
		logger:        logger,
		retryAttempts: retryAttempts,
	}, nil
}

func (s *DBStorage) Close() {
	err := s.db.Close()
	if err != nil {
		s.logger.Error("failed to close database", zap.Error(err))
	}
}

func (s *DBStorage) WithTransaction(ctx context.Context) (context.Context, *sql.Tx, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, nil, fmt.Errorf("transaction begin failed: %w", err)
	}
	ctxWithTransaction := context.WithValue(ctx, transactionKey, tx)
	return ctxWithTransaction, tx, nil
}

func (s *DBStorage) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	tx, err := getTransaction(ctx)
	if err != nil {
		return nil, err
	}
	return tx.ExecContext(ctx, query, args...) //nolint:wrapcheck // unnecessary
}

func (s *DBStorage) QueryRow(ctx context.Context, query string, args ...any) (*sql.Row, error) {
	tx, err := getTransaction(ctx)
	if err != nil {
		return nil, err
	}
	return tx.QueryRowContext(ctx, query, args...), nil
}

func (s *DBStorage) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	tx, err := getTransaction(ctx)
	if err != nil {
		return nil, err
	}
	return tx.QueryContext(ctx, query, args...) //nolint:wrapcheck // unnecessary
}

func getTransaction(ctx context.Context) (*sql.Tx, error) {
	txVal := ctx.Value(transactionKey)
	if txVal == nil {
		return nil, errors.New("no transaction set")
	}
	tx, ok := txVal.(*sql.Tx)
	if !ok {
		return nil, errors.New("invalid transaction type")
	}
	return tx, nil
}
