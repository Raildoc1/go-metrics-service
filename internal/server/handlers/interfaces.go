package handlers

import (
	"context"
	"errors"
	"go-metrics-service/internal/common/protocol"
)

// Errors.

var (
	ErrNonExistentType = errors.New("non-existent type")
	ErrWrongValueType  = errors.New("wrong value type")
	ErrParsing         = errors.New("parsing error")
)

// Data.

type GaugeRepository interface {
	GetGauge(ctx context.Context, key string) (float64, error)
}

type CounterRepository interface {
	GetCounter(ctx context.Context, key string) (int64, error)
}

type AllMetricsRepository interface {
	GetAll(ctx context.Context) (map[string]any, error)
}

// Logic.

type MetricController interface {
	Update(ctx context.Context, metric protocol.Metrics) error
	UpdateMany(ctx context.Context, metrics []protocol.Metrics) error
}
