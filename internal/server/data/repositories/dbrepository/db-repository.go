package dbrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

type DBStorage interface {
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) (*sql.Row, error)
}

type DBRepository struct {
	storage DBStorage
	logger  *zap.Logger
}

const (
	dbQueryFailedMsg = "database query failed: %w"
)

func New(storage DBStorage, logger *zap.Logger) *DBRepository {
	return &DBRepository{
		storage: storage,
		logger:  logger,
	}
}

func (r *DBRepository) Has(ctx context.Context, key string) (bool, error) {
	res, err := getValue[int](ctx, r.storage, "count(1)", key, 0)
	if err != nil {
		return false, fmt.Errorf(dbQueryFailedMsg, err)
	}
	return res > 0, nil
}

func (r *DBRepository) GetCounter(ctx context.Context, key string) (int64, error) {
	return getValue[int64](ctx, r.storage, "counter_value", key, 0)
}

func (r *DBRepository) GetGauge(ctx context.Context, key string) (float64, error) {
	return getValue[float64](ctx, r.storage, "gauge_value", key, 0)
}

func getValue[T any](
	ctx context.Context,
	storage DBStorage,
	dbFieldName, key string,
	defaultVal T,
) (T, error) {
	query := fmt.Sprintf(`
		select %s from metrics
		where key=$1
	`, dbFieldName)
	row, err := storage.QueryRow(ctx, query, key)
	if err != nil {
		return defaultVal, fmt.Errorf(dbQueryFailedMsg, err)
	}
	if err := row.Err(); err != nil {
		return defaultVal, fmt.Errorf(dbQueryFailedMsg, err)
	}
	var c T
	err = row.Scan(&c)
	switch {
	case err == nil:
		return c, nil
	case errors.Is(err, sql.ErrNoRows):
		return defaultVal, nil
	default:
		return defaultVal, fmt.Errorf(dbQueryFailedMsg, err)
	}
}

func (r *DBRepository) SetCounter(ctx context.Context, key string, value int64) error {
	const query = `
		insert into metrics (key, counter_value)
		values ($1, $2)
		on conflict (key)
			do update set counter_value = $2;`
	_, err := r.storage.Exec(ctx, query, key, value)
	if err != nil {
		return fmt.Errorf("setting counter failed: %w", err)
	}
	return nil
}

func (r *DBRepository) SetCounters(ctx context.Context, values map[string]int64) error {
	genericValues := make(map[string]any, len(values))
	for key, value := range values {
		genericValues[key] = value
	}
	return r.setMany(ctx, "counter_value", genericValues)
}

func (r *DBRepository) SetGauges(ctx context.Context, values map[string]float64) error {
	genericValues := make(map[string]any, len(values))
	for key, value := range values {
		genericValues[key] = value
	}
	return r.setMany(ctx, "gauge_value", genericValues)
}

func (r *DBRepository) setMany(ctx context.Context, dbFieldName string, values map[string]any) error {
	if len(values) == 0 {
		return nil
	}
	const queryPattern = `
		insert into metrics (key, %s)
		values %s
		on conflict (key)
		    do update set %s = excluded.%s`
	const firstArgNumber = 1
	const argsIsRow = 2
	query := fmt.Sprintf(
		queryPattern,
		dbFieldName,
		formatValuesRows(firstArgNumber, argsIsRow, len(values)),
		dbFieldName,
		dbFieldName,
	)
	args := make([]any, 0, len(values)*2)
	for key, value := range values {
		args = append(args, key, value)
	}
	_, err := r.storage.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("setting values failed: %w", err)
	}
	return nil
}

func (r *DBRepository) SetGauge(ctx context.Context, key string, value float64) error {
	const query = `
		insert into metrics (key, gauge_value)
		values ($1, $2)
		on conflict (key)
			do update set gauge_value = $2;`
	_, err := r.storage.Exec(ctx, query, key, value)
	if err != nil {
		return fmt.Errorf("setting gauge failed: %w", err)
	}
	return nil
}

func (r *DBRepository) GetAll(ctx context.Context) (map[string]any, error) {
	query := `select key, gauge_value, counter_value from metrics`
	rows, err := r.storage.Query(ctx, query) //nolint:sqlclosecheck // rows are closed below
	if err != nil {
		return nil, fmt.Errorf(dbQueryFailedMsg, err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			r.logger.Error("failed to close database rows", zap.Error(err))
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
			r.logger.Error("null value read", zap.String("key", m.key))
		}
	}
	return res, nil
}

func formatValuesRows(firstNumber, valuesCount, rowsCount int) string {
	currentNum := firstNumber
	rows := make([]string, rowsCount)
	for i := range rowsCount {
		rows[i] = formatRow(currentNum, valuesCount)
		currentNum += valuesCount
	}
	return strings.Join(rows, ",")
}

func formatRow(firstNumber, valuesCount int) string {
	currentNum := firstNumber
	values := make([]string, valuesCount)
	for i := range valuesCount {
		values[i] = fmt.Sprintf("$%v", currentNum)
		currentNum++
	}
	return fmt.Sprintf("(%s)", strings.Join(values, ","))
}
