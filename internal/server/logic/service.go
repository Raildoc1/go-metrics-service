// Package logic contains logic
package logic

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type Repository interface {
	Has(ctx context.Context, key string) (bool, error)
	GetCounter(ctx context.Context, key string) (int64, error)
	GetGauge(ctx context.Context, key string) (float64, error)
	SetCounter(ctx context.Context, key string, value int64) error
	SetCounters(ctx context.Context, values map[string]int64) error
	SetGauge(ctx context.Context, key string, value float64) error
	SetGauges(ctx context.Context, values map[string]float64) error
	GetAll(ctx context.Context) (map[string]any, error)
}

type Service struct {
	r Repository
	l *zap.Logger
}

type CounterDiff struct {
	Key   string
	Delta int64
}

type GaugeDiff struct {
	Key      string
	NewValue float64
}

func NewService(r Repository, l *zap.Logger) *Service {
	return &Service{
		r: r,
		l: l,
	}
}

func (s *Service) UpdateGauge(ctx context.Context, diff GaugeDiff) error {
	err := s.r.SetGauge(ctx, diff.Key, diff.NewValue)
	if err != nil {
		return fmt.Errorf("failed to set gauge: %w", err)
	}
	return nil
}

func (s *Service) UpdateGauges(ctx context.Context, diffs []GaugeDiff) error {
	values := make(map[string]float64)
	for _, diff := range diffs {
		values[diff.Key] = diff.NewValue
	}
	err := s.r.SetGauges(ctx, values)
	if err != nil {
		return fmt.Errorf("failed to set gauges: %w", err)
	}
	return nil
}

func (s *Service) UpdateCounter(ctx context.Context, diff CounterDiff) error {
	newValue, err := s.getChangedCounter(ctx, diff.Key, diff.Delta)
	if err != nil {
		return err
	}
	err = s.r.SetCounter(ctx, diff.Key, newValue)
	if err != nil {
		return fmt.Errorf("failed to set counter: %w", err)
	}
	return nil
}

func (s *Service) UpdateCounters(ctx context.Context, diffs []CounterDiff) error {
	values := make(map[string]int64)
	for _, diff := range diffs {
		if _, ok := values[diff.Key]; ok {
			values[diff.Key] += diff.Delta
		} else {
			newValue, err := s.getChangedCounter(ctx, diff.Key, diff.Delta)
			if err != nil {
				return err
			}
			values[diff.Key] = newValue
		}
	}
	err := s.r.SetCounters(ctx, values)
	if err != nil {
		return fmt.Errorf("failed to set counters: %w", err)
	}
	return nil
}

func (s *Service) getChangedCounter(ctx context.Context, key string, delta int64) (int64, error) {
	hasValue, err := s.r.Has(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("hasValue: %w", err)
	}
	var prevValue int64
	if !hasValue {
		prevValue = int64(0)
	} else {
		var err error
		prevValue, err = s.r.GetCounter(ctx, key)
		if err != nil {
			return 0, fmt.Errorf("%w: getting counter '%s' failed", err, key)
		}
	}
	newValue := prevValue + delta
	s.l.Debug(
		"change counter",
		zap.String("key", key),
		zap.Int64("value", newValue),
		zap.Int64("delta", delta),
		zap.Int64("prevValue", prevValue),
	)
	return newValue, nil
}
