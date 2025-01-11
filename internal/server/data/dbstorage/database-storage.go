package dbstorage

import (
	"database/sql"
	"errors"
	"fmt"
	"go-metrics-service/internal/server/data"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	setupDatabaseRequest = `
		create table if not exists metrics
		(
			key           varchar(63) not null primary key,
			gauge_value   double precision,
			counter_value integer
		);
		
		create or replace function fn_validate_metric() returns trigger as
		'
			begin
				if (new.counter_value is null) = (new.gauge_value is null) then
					raise exception ''Either counter_value or gauge_value must be null''
						using errcode = ''MV1A0'';
				end if;
				return new;
			end;
		' language plpgsql;
		
		create or replace trigger before_insert_metric_trigger
			before insert
			on metrics
			for each row
		execute function fn_validate_metric();
		
		create or replace trigger before_update_metric_trigger
			before update
			on metrics
			for each row
		execute function fn_validate_metric();`

	hasMetricRequest = `
		select count(1) from metrics
		where key=$1
	`

	getCounterRequest = `
		select counter_value from metrics
		where key=$1
	`

	getGaugeRequest = `
		select gauge_value from metrics
		where key=$1
	`

	getAllRequest = `select * from metrics`
)

const (
	dbQueryFailedMsg = "database query failed: %w"
)

type DBFactory interface {
	Create() (*sql.DB, error)
}

type DBStorage struct {
	transaction      *transaction
	db               *sql.DB
	logger           *zap.Logger
	transactionMutex sync.Mutex
}

type transaction struct {
	tx *sql.Tx
	id data.TransactionID
}

func New(dbFactory DBFactory, logger *zap.Logger) (*DBStorage, error) {
	db, err := dbFactory.Create()
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}
	_, err = db.Exec(setupDatabaseRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}
	return &DBStorage{
		transaction: nil,
		db:          db,
		logger:      logger,
	}, nil
}

func (s *DBStorage) BeginTransaction() (data.TransactionID, error) {
	s.transactionMutex.Lock()
	if s.transaction != nil {
		return "", errors.New("transaction is already opened")
	}
	id := uuid.New()
	transactionID := data.TransactionID(id.String())
	tx, err := s.db.Begin()
	if err != nil {
		s.transactionMutex.Unlock()
		return "", fmt.Errorf("failed to start transaction: %w", err)
	}
	s.transaction = &transaction{
		id: transactionID,
		tx: tx,
	}
	return transactionID, nil
}

func (s *DBStorage) CommitTransaction(transactionID data.TransactionID) error {
	if s.transaction == nil {
		return errors.New("no transaction opened")
	}
	if s.transaction.id != transactionID {
		return errors.New("wrong transaction id")
	}

	defer func(s *DBStorage) {
		err := s.transaction.tx.Rollback()
		switch {
		case err == nil:
		case errors.Is(err, sql.ErrTxDone):
			// Ignore.
		default:
			s.logger.Error("failed to rollback transaction", zap.Error(err))
		}
		s.transaction = nil
		s.transactionMutex.Unlock()
	}(s)

	err := s.transaction.tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (s *DBStorage) RollbackTransaction(transactionID data.TransactionID) error {
	if s.transaction == nil {
		return errors.New("no transaction opened")
	}
	if s.transaction.id != transactionID {
		return data.ErrWrongTransactionID
	}

	defer func(s *DBStorage) {
		s.transaction = nil
		s.transactionMutex.Unlock()
	}(s)

	err := s.transaction.tx.Rollback()
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	return nil
}

func (s *DBStorage) SetCounter(key string, value int64, transactionID data.TransactionID) error {
	if err := s.validateTransactionID(transactionID); err != nil {
		return err
	}
	const upsertCounterRequest = `
		insert into metrics (key, counter_value)
		values ($1, $2)
		on conflict (key)
			do update set counter_value = $2;`
	_, err := s.transaction.tx.Exec(upsertCounterRequest, key, value)
	if err != nil {
		return fmt.Errorf("setting counter failed: %w", err)
	}
	return nil
}

func (s *DBStorage) SetGauge(key string, value float64, transactionID data.TransactionID) error {
	if err := s.validateTransactionID(transactionID); err != nil {
		return err
	}
	const upsertGaugeRequest = `
		insert into metrics (key, gauge_value)
		values ($1, $2)
		on conflict (key)
			do update set gauge_value = $2;`
	_, err := s.transaction.tx.Exec(upsertGaugeRequest, key, value)
	if err != nil {
		return fmt.Errorf("setting gauge failed: %w", err)
	}
	return nil
}

func (s *DBStorage) Has(key string) (bool, error) {
	row := s.db.QueryRow(hasMetricRequest, key)
	if err := row.Err(); err != nil {
		return false, fmt.Errorf(dbQueryFailedMsg, err)
	}
	var res int
	if err := row.Scan(&res); err != nil {
		return false, fmt.Errorf(dbQueryFailedMsg, err)
	}
	return res > 0, nil
}

func (s *DBStorage) GetCounter(key string) (int64, error) {
	row := s.db.QueryRow(getCounterRequest, key)
	if err := row.Err(); err != nil {
		return 0, fmt.Errorf(dbQueryFailedMsg, err)
	}
	var c int64
	if err := row.Scan(&c); err != nil {
		return 0, fmt.Errorf(dbQueryFailedMsg, err)
	}
	return c, nil
}

func (s *DBStorage) GetGauge(key string) (float64, error) {
	row := s.db.QueryRow(getGaugeRequest, key)
	if err := row.Err(); err != nil {
		return 0, fmt.Errorf(dbQueryFailedMsg, err)
	}
	var g float64
	if err := row.Scan(&g); err != nil {
		return 0, fmt.Errorf(dbQueryFailedMsg, err)
	}
	return g, nil
}

func (s *DBStorage) GetAll() (map[string]any, error) {
	rows, err := s.db.Query(getAllRequest) //nolint:sqlclosecheck // rows are closed below
	if err != nil {
		return nil, fmt.Errorf(dbQueryFailedMsg, err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			s.logger.Error("failed to close database rows", zap.Error(err))
		}
	}(rows)
	if rows.Err() != nil {
		return nil, fmt.Errorf(dbQueryFailedMsg, rows.Err())
	}
	type metric struct {
		gaugeValue   *float64
		counterValue *int64
		key          string
	}
	res := make(map[string]any)
	for rows.Next() {
		var m metric
		if err := rows.Scan(&m.key, &m.gaugeValue, &m.counterValue); err != nil {
			return nil, fmt.Errorf(dbQueryFailedMsg, err)
		}
		switch {
		case m.counterValue != nil:
			res[m.key] = *m.counterValue
		case m.gaugeValue != nil:
			res[m.key] = *m.gaugeValue
		default:
			s.logger.Error("null value read", zap.String("key", m.key))
		}
	}
	return res, nil
}

func (s *DBStorage) Close() {
	if s.transaction != nil {
		err := s.transaction.tx.Rollback()
		if err != nil {
			s.logger.Error("failed to rollback transaction", zap.Error(err))
		}
		s.transaction = nil
		s.transactionMutex.Unlock()
	}
	err := s.db.Close()
	if err != nil {
		s.logger.Error("failed to close database", zap.Error(err))
	}
}

func (s *DBStorage) validateTransactionID(transactionID data.TransactionID) error {
	if s.transaction == nil || s.transaction.id != transactionID {
		return data.ErrWrongTransactionID
	}
	return nil
}
