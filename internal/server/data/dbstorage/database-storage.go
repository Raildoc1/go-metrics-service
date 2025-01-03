package dbstorage

import (
	"database/sql"
	"fmt"

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

	upsertCounterRequest = `
		insert into metrics (key, counter_value)
		values ($1, $2)
		on conflict (key)
			do update set counter_value = $2;`

	upsertGaugeRequest = `
		insert into metrics (key, gauge_value)
		values ($1, $2)
		on conflict (key)
			do update set gauge_value = $2;`

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

type DBFactory interface {
	Create() (*sql.DB, error)
}

type DBStorage struct {
	db     *sql.DB
	logger *zap.Logger
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
		db:     db,
		logger: logger,
	}, nil
}

func (s *DBStorage) SetCounter(key string, value int64) error {
	_, err := s.db.Exec(upsertCounterRequest, key, value)
	if err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}
	return nil
}

func (s *DBStorage) SetGauge(key string, value float64) error {
	_, err := s.db.Exec(upsertGaugeRequest, key, value)
	if err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}
	return nil
}

func (s *DBStorage) Has(key string) (bool, error) {
	row := s.db.QueryRow(hasMetricRequest, key)
	if err := row.Err(); err != nil {
		return false, fmt.Errorf("database query failed: %w", err)
	}
	var res int
	if err := row.Scan(&res); err != nil {
		return false, fmt.Errorf("database query failed: %w", err)
	}
	return res > 0, nil
}

func (s *DBStorage) GetCounter(key string) (int64, error) {
	row := s.db.QueryRow(getCounterRequest, key)
	if err := row.Err(); err != nil {
		return 0, fmt.Errorf("database query failed: %w", err)
	}
	var c int64
	if err := row.Scan(&c); err != nil {
		return 0, fmt.Errorf("database query failed: %w", err)
	}
	return c, nil
}

func (s *DBStorage) GetGauge(key string) (float64, error) {
	row := s.db.QueryRow(getGaugeRequest, key)
	if err := row.Err(); err != nil {
		return 0, fmt.Errorf("database query failed: %w", err)
	}
	var g float64
	if err := row.Scan(&g); err != nil {
		return 0, fmt.Errorf("database query failed: %w", err)
	}
	return g, nil
}

func (s *DBStorage) GetAll() (map[string]any, error) {
	rows, err := s.db.Query(getAllRequest)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			s.logger.Error("failed to close database rows", zap.Error(err))
		}
	}(rows)
	type metric struct {
		key          string
		gaugeValue   *float64
		counterValue *int64
	}
	res := make(map[string]any)
	for rows.Next() {
		var m metric
		if err := rows.Scan(&m.key, &m.gaugeValue, &m.counterValue); err != nil {
			return nil, fmt.Errorf("database query failed: %w", err)
		}
		if m.counterValue != nil {
			res[m.key] = *m.counterValue
		} else if m.gaugeValue != nil {
			res[m.key] = *m.gaugeValue
		} else {
			s.logger.Error("null value read", zap.String("key", m.key))
		}
	}
	return res, nil
}

func (s *DBStorage) Close() {
	err := s.db.Close()
	s.logger.Error("failed to close database", zap.Error(err))
}
