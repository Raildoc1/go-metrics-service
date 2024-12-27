package logic

import (
	"fmt"

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

func (g *Gauge) Set(key string, value float64) error {
	g.logger.Debug("Setting", zap.String("key", key), zap.Float64("value", value))
	err := g.repository.SetFloat64(key, value)
	if err != nil {
		return fmt.Errorf("%w: setting gauge '%s' failed", err, key)
	}
	return nil
}
