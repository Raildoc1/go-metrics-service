package handlers

import (
	"context"
	"errors"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data"
)

// Errors.

var (
	ErrNonExistentType = errors.New("non-existent type")
	ErrWrongValueType  = errors.New("wrong value type")
	ErrParsing         = errors.New("parsing error")
)

// Data.

type GaugeRepository interface {
	SetGauge(ctx context.Context, key string, value float64, transactionID data.TransactionID) error
	GetGauge(ctx context.Context, key string) (float64, error)
}

type CounterRepository interface {
	SetCounter(ctx context.Context, key string, value int64, transactionID data.TransactionID) error
	GetCounter(ctx context.Context, key string) (int64, error)
}

type AllMetricsRepository interface {
	GetAll(ctx context.Context) (map[string]any, error)
}

// Logic.

type MetricUpdater interface {
	UpdateOne(ctx context.Context, metric protocol.Metrics) error
	UpdateMany(ctx context.Context, metrics []protocol.Metrics) error
}

type CounterLogic interface {
	Change(ctx context.Context, key string, delta int64, transactionID data.TransactionID) error
}

type GaugeLogic interface {
	Set(ctx context.Context, key string, value float64, transactionID data.TransactionID) error
}
