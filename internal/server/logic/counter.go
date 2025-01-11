package logic

import (
	"fmt"
	"go-metrics-service/internal/server/data"

	"go.uber.org/zap"
)

type Counter struct {
	repository CounterRepository
	logger     *zap.Logger
}

func NewCounter(repository CounterRepository, logger *zap.Logger) *Counter {
	return &Counter{
		repository: repository,
		logger:     logger,
	}
}

func (c *Counter) Change(key string, delta int64, transactionID data.TransactionID) error {
	c.logger.Debug("changing", zap.String("key", key), zap.Int64("delta", delta))
	hasValue, err := c.repository.Has(key)
	if err != nil {
		return fmt.Errorf("hasValue: %w", err)
	}
	var prevValue int64
	if !hasValue {
		prevValue = int64(0)
	} else {
		var err error
		prevValue, err = c.repository.GetCounter(key)
		if err != nil {
			return fmt.Errorf("%w: getting counter '%s' failed", err, key)
		}
	}
	newValue := prevValue + delta
	err = c.repository.SetCounter(key, newValue, transactionID)
	if err != nil {
		return fmt.Errorf("%w: setting counter '%s' failed", err, key)
	}
	return nil
}
