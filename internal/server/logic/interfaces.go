package logic

import (
	"context"
	"go-metrics-service/internal/server/data"
)

type CounterRepository interface {
	Has(ctx context.Context, key string) (bool, error)
	SetCounter(ctx context.Context, key string, value int64, transactionID data.TransactionID) error
	GetCounter(ctx context.Context, key string) (int64, error)
}

type GaugeRepository interface {
	SetGauge(ctx context.Context, key string, value float64, transactionID data.TransactionID) error
}
