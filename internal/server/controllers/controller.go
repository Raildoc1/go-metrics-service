package controllers

import (
	"context"
	"errors"
	"fmt"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/logic"

	"go.uber.org/zap"
)

type Controller struct {
	l  *zap.Logger
	tm TransactionManager
	s  Service
}

type Service interface {
	UpdateGauge(ctx context.Context, diff logic.GaugeDiff) error
	UpdateGauges(ctx context.Context, diffs []logic.GaugeDiff) error
	UpdateCounter(ctx context.Context, diff logic.CounterDiff) error
	UpdateCounters(ctx context.Context, diffs []logic.CounterDiff) error
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
			if err := c.s.UpdateGauge(
				ctx,
				logic.GaugeDiff{
					Key:      metric.ID,
					NewValue: *metric.Value,
				},
			); err != nil {
				return fmt.Errorf("set gauge: %w", err)
			}
		case protocol.Counter:
			if metric.Delta == nil {
				return ErrWrongValueType
			}
			if err := c.s.UpdateCounter(
				ctx,
				logic.CounterDiff{
					Key:   metric.ID,
					Delta: *metric.Delta,
				},
			); err != nil {
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
		counterDiffs := make([]logic.CounterDiff, 0)
		gaugeDiffs := make([]logic.GaugeDiff, 0)
		for _, metric := range metrics {
			switch metric.MType {
			case protocol.Gauge:
				if metric.Value == nil {
					return ErrWrongValueType
				}
				gaugeDiffs = append(
					gaugeDiffs,
					logic.GaugeDiff{
						Key:      metric.ID,
						NewValue: *metric.Value,
					},
				)
			case protocol.Counter:
				if metric.Delta == nil {
					return ErrWrongValueType
				}
				counterDiffs = append(
					counterDiffs,
					logic.CounterDiff{
						Key:   metric.ID,
						Delta: *metric.Delta,
					},
				)
			default:
				return ErrNonExistentType
			}
		}
		err := c.s.UpdateCounters(ctx, counterDiffs)
		if err != nil {
			return fmt.Errorf("update counters failed: %w", err)
		}
		err = c.s.UpdateGauges(ctx, gaugeDiffs)
		if err != nil {
			return fmt.Errorf("update gauges failed: %w", err)
		}
		return nil
	})
}

func (c *Controller) SetGauge(ctx context.Context, key string, value float64) error {
	c.l.Debug("changing", zap.String("key", key), zap.Float64("value", value))
	return c.tm.DoWithTransaction(ctx, func(ctx context.Context) error { //nolint:wrapcheck // unnecessary
		return c.s.UpdateGauge(
			ctx,
			logic.GaugeDiff{
				Key:      key,
				NewValue: value,
			},
		)
	})
}

func (c *Controller) ChangeCounter(ctx context.Context, key string, delta int64) error {
	c.l.Debug("changing", zap.String("key", key), zap.Int64("delta", delta))
	return c.tm.DoWithTransaction(ctx, func(ctx context.Context) error { //nolint:wrapcheck // unnecessary
		return c.s.UpdateCounter(
			ctx,
			logic.CounterDiff{
				Key:   key,
				Delta: delta,
			},
		)
	})
}
