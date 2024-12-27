package logic

import (
	"fmt"

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

func (c *Counter) Change(key string, delta int64) error {
	c.logger.Debug("changing", zap.String("key", key), zap.Int64("delta", delta))
	var prevValue int64
	if !c.repository.Has(key) {
		prevValue = int64(0)
	} else {
		var err error
		prevValue, err = c.repository.GetInt64(key)
		if err != nil {
			return fmt.Errorf("%w: getting counter '%s' failed", err, key)
		}
	}
	newValue := prevValue + delta
	err := c.repository.SetInt64(key, newValue)
	if err != nil {
		return fmt.Errorf("%w: setting counter '%s' failed", err, key)
	}
	return nil
}
