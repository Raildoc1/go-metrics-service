package handlers

import "errors"

// Errors

var (
	ErrNonExistentType = errors.New("non-existent type")
)

// Data

type GaugeRepository interface {
	SetFloat64(key string, value float64) error
	GetFloat64(key string) (value float64, err error)
}

type CounterRepository interface {
	SetInt64(key string, value int64) error
	GetInt64(key string) (value int64, err error)
}

type AllMetricsRepository interface {
	GetAll() map[string]any
}

// Logic

type CounterLogic interface {
	Change(key string, delta int64) error
}

type GaugeLogic interface {
	Set(key string, value float64) error
}

// Utils

type Logger interface {
	Errorln(args ...interface{})
	Debugln(args ...interface{})
}