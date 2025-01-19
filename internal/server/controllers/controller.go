package controllers

import (
	"context"
	"errors"
	"fmt"
	"go-metrics-service/internal/common/protocol"

	"go.uber.org/zap"
)

type Controller struct {
	l  *zap.Logger
	tm TransactionManager
	s  Service
}

type Service interface {
	SetGauge(ctx context.Context, key string, value float64) error
	ChangeCounter(ctx context.Context, key string, delta int64) error
	HandleMany(ctx context.Context, values map[string]any) error
}

type TransactionManager interface {
	DoWithTransaction(ctx context.Context, f func(ctx context.Context) error) error
}

var (
	ErrNonExistentType = errors.New("non-existent type")
	ErrWrongValueType  = errors.New("wrong value type")
)

func NewController(tm TransactionManager, gs Service, l *zap.Logger) *Controller {
	return &Controller{
		l:  l,
		tm: tm,
		s:  gs,
	}
}

func (c *Controller) Update(ctx context.Context, metric protocol.Metrics) error {
	return c.tm.DoWithTransaction(ctx, func(ctx context.Context) error { //nolint:wrapcheck // unnecessary
		switch metric.MType {
		case protocol.Gauge:
			if metric.Value == nil {
				return ErrWrongValueType
			}
			if err := c.s.SetGauge(ctx, metric.ID, *metric.Value); err != nil {
				return fmt.Errorf("set gauge: %w", err)
			}
		case protocol.Counter:
			if metric.Delta == nil {
				return ErrWrongValueType
			}
			if err := c.s.ChangeCounter(ctx, metric.ID, *metric.Delta); err != nil {
				return fmt.Errorf("change counter: %w", err)
			}
		default:
			return ErrNonExistentType
		}
		return nil
	})
}

func (c *Controller) UpdateMany(ctx context.Context, metrics []protocol.Metrics) error {
	return c.tm.DoWithTransaction(ctx, func(ctx context.Context) error { //nolint:wrapcheck // unnecessary
		values := make(map[string]any, len(metrics))
		for _, metric := range metrics {
			switch metric.MType {
			case protocol.Gauge:
				if metric.Value == nil {
					return ErrWrongValueType
				}
				values[metric.ID] = *metric.Value
			case protocol.Counter:
				if metric.Delta == nil {
					return ErrWrongValueType
				}
				values[metric.ID] = *metric.Delta
			default:
				return ErrNonExistentType
			}
		}
		return c.s.HandleMany(ctx, values)
	})
}

func (c *Controller) SetGauge(ctx context.Context, key string, value float64) error {
	c.l.Debug("changing", zap.String("key", key), zap.Float64("value", value))
	return c.tm.DoWithTransaction(ctx, func(ctx context.Context) error { //nolint:wrapcheck // unnecessary
		return c.s.SetGauge(ctx, key, value)
	})
}

func (c *Controller) ChangeCounter(ctx context.Context, key string, delta int64) error {
	c.l.Debug("changing", zap.String("key", key), zap.Int64("delta", delta))
	return c.tm.DoWithTransaction(ctx, func(ctx context.Context) error { //nolint:wrapcheck // unnecessary
		return c.s.ChangeCounter(ctx, key, delta)
	})
}

func (c *Controller) HandleMany(ctx context.Context, values map[string]any) error {
	return c.tm.DoWithTransaction(ctx, func(ctx context.Context) error { //nolint:wrapcheck // unnecessary
		return c.s.HandleMany(ctx, values)
	})
}
