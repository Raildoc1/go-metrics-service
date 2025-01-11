package logic

import "go-metrics-service/internal/server/data"

type CounterRepository interface {
	Has(key string) (bool, error)
	SetCounter(key string, value int64, transactionID data.TransactionID) error
	GetCounter(key string) (int64, error)
}

type GaugeRepository interface {
	SetGauge(key string, value float64, transactionID data.TransactionID) error
}
