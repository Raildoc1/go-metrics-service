package handlers

import "errors"

// Errors.

var (
	ErrNonExistentType = errors.New("non-existent type")
	ErrWrongValueType  = errors.New("wrong value type")
	ErrParsing         = errors.New("parsing error")
)

// Data.

type GaugeRepository interface {
	SetGauge(key string, value float64) error
	GetGauge(key string) (float64, error)
}

type CounterRepository interface {
	SetCounter(key string, value int64) error
	GetCounter(key string) (int64, error)
}

type AllMetricsRepository interface {
	GetAll() (map[string]any, error)
}

// Logic.

type CounterLogic interface {
	Change(key string, delta int64) error
}

type GaugeLogic interface {
	Set(key string, value float64) error
}
