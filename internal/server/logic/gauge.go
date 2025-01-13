package logic

import (
	"context"
	"fmt"
	"go-metrics-service/internal/server/data"

	"go.uber.org/zap"
)

type Gauge struct {
	repository GaugeRepository
	logger     *zap.Logger
}

func New(repository GaugeRepository, logger *zap.Logger) *Gauge {
	return &Gauge{
		repository: repository,
		logger:     logger,
	}
}

func (g *Gauge) Set(ctx context.Context, key string, value float64, transactionID data.TransactionID) error {
	g.logger.Debug("Setting", zap.String("key", key), zap.Float64("value", value))
	err := g.repository.SetGauge(ctx, key, value, transactionID)
	if err != nil {
		return fmt.Errorf("%w: setting gauge '%s' failed", err, key)
	}
	return nil
}
