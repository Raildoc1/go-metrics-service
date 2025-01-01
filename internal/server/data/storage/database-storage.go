package storage

import (
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

const (
	setupDatabaseRequest = `
		create table if not exists metrics (
			key varchar(63) not null primary key,
			gauge_value double precision,
			counter_value integer
		);

		create or replace function fn_before_insert_metric() returns trigger as
		'
			begin
				if (new.counter_value is null) = (new.gauge_value is null) then
					raise exception ''Either counter_value or gauge_value must be null''
						using errcode = ''MI1A0'';
				end if;
				return new;
			end;
		' language plpgsql;

		create or replace function fn_before_update_counter() returns trigger as
		'
			begin
				if old.gauge_value is not null then
					raise exception ''Cannot update counter_value of gauge''
						using errcode = ''MU1C0'';
				end if;
				return new;
			end;
		' language plpgsql;
		
		create or replace function fn_before_update_gauge() returns trigger as
		'
			begin
				if old.counter_value is not null then
					raise exception ''Cannot update gauge_value of counter''
						using errcode = ''MU1G0'';
				end if;
				return new;
			end;
		' language plpgsql;
		
		create or replace trigger before_insert_metric_trigger
			before insert
			on metrics
			for each row
		execute function fn_before_insert_metric();
		
		create or replace trigger before_update_counter_trigger
			before update of counter_value
			on metrics
			for each row
		execute function fn_before_update_counter();
		
		create or replace trigger before_update_gauge_trigger
			before update of gauge_value
			on metrics
			for each row
		execute function fn_before_update_gauge();`

	insertGaugeRequest = `INSERT INTO metrics (key, gauge_value) VALUES (?, ?)`
	updateGaugeRequest = `
UPDATE metrics SET gauge_value = ? WHERE key = ?
`
	insertCounterRequest = `INSERT INTO metrics (key, counter_value) VALUES (?, ?)`
	updateCounterRequest = `UPDATE metrics SET counter_value = ? WHERE key = ?`
)

type DatabaseStorage struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewDatabaseStorage(db *sql.DB, logger *zap.Logger) (*DatabaseStorage, error) {
	_, err := db.Exec(setupDatabaseRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create database storage: %w", err)
	}
	return &DatabaseStorage{
		db:     db,
		logger: logger,
	}, nil
}

func (s *DatabaseStorage) SetGauge(key string, value int64) {

}

func (s *DatabaseStorage) Has(key string) bool {
	_, ok := m.data[key]
	return ok
}

func (s *DatabaseStorage) Get(key string) (any, bool) {
	v, ok := m.data[key]
	return v, ok
}

func (s *DatabaseStorage) GetAll() map[string]any {
	return m.data
}
