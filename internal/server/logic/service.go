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

func NewService(r Repository, l *zap.Logger) *Service {
	return &Service{
		r: r,
		l: l,
	}
}

func (s *Service) HandleMany(ctx context.Context, values map[string]any) error {
	gauges := make(map[string]float64)
	counters := make(map[string]int64)
	for k, v := range values {
		switch casted := v.(type) {
		case float64:
			gauges[k] = casted
		case int64:
			counters[k] = casted
		default:
			return fmt.Errorf("invalid value type: %T", casted)
		}
	}
	err := s.SetGauges(ctx, gauges)
	if err != nil {
		return err
	}
	err = s.ChangeCounters(ctx, counters)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) SetGauge(ctx context.Context, key string, value float64) error {
	err := s.r.SetGauge(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to set gauge: %w", err)
	}
	return nil
}

func (s *Service) SetGauges(ctx context.Context, values map[string]float64) error {
	err := s.r.SetGauges(ctx, values)
	if err != nil {
		return fmt.Errorf("failed to set gauges: %w", err)
	}
	return nil
}

func (s *Service) ChangeCounter(ctx context.Context, key string, delta int64) error {
	newValue, err := s.getChangedCounter(ctx, key, delta)
	if err != nil {
		return err
	}
	err = s.r.SetCounter(ctx, key, newValue)
	if err != nil {
		return fmt.Errorf("failed to set counter: %w", err)
	}
	return nil
}

func (s *Service) ChangeCounters(ctx context.Context, deltas map[string]int64) error {
	values := make(map[string]int64, len(deltas))
	for k, d := range deltas {
		newValue, err := s.getChangedCounter(ctx, k, d)
		if err != nil {
			return err
		}
		values[k] = newValue
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
