package handlers

import (
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
	SetGauge(key string, value float64, transactionID data.TransactionID) error
	GetGauge(key string) (float64, error)
}

type CounterRepository interface {
	SetCounter(key string, value int64, transactionID data.TransactionID) error
	GetCounter(key string) (int64, error)
}

type AllMetricsRepository interface {
	GetAll() (map[string]any, error)
}

// Logic.

type MetricUpdater interface {
	UpdateOne(metric protocol.Metrics) error
	UpdateMany(metrics []protocol.Metrics) error
}

type CounterLogic interface {
	Change(key string, delta int64, transactionID data.TransactionID) error
}

type GaugeLogic interface {
	Set(key string, value float64, transactionID data.TransactionID) error
}
