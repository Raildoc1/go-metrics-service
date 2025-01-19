package memrepository

import (
	"context"
	"go-metrics-service/internal/server/data"

	"go.uber.org/zap"
)

type MemStorage interface {
	Get(key string) (val any, ok bool)
	GetAll() map[string]any
	Set(key string, value any)
}

type MemRepository struct {
	storage MemStorage
	logger  *zap.Logger
}

func (r *MemRepository) SetCounters(ctx context.Context, values map[string]int64) error {
	for k, v := range values {
		err := r.SetCounter(ctx, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MemRepository) SetGauges(ctx context.Context, values map[string]float64) error {
	for k, v := range values {
		err := r.SetGauge(ctx, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func New(storage MemStorage, logger *zap.Logger) *MemRepository {
	return &MemRepository{
		storage: storage,
		logger:  logger,
	}
}

func (r *MemRepository) Has(_ context.Context, key string) (bool, error) {
	_, ok := r.storage.Get(key)
	return ok, nil
}

func (r *MemRepository) GetCounter(_ context.Context, key string) (int64, error) {
	return getInternal[int64](r, key, 0)
}

func (r *MemRepository) GetGauge(_ context.Context, key string) (float64, error) {
	return getInternal[float64](r, key, 0.0)
}

func getInternal[T any](r *MemRepository, key string, defaultValue T) (T, error) {
	val, ok := r.storage.Get(key)
	if !ok {
		return defaultValue, data.ErrNotFound
	}
	res, ok := val.(T)
	if !ok {
		return defaultValue, data.ErrWrongType
	}
	return res, nil
}

func (r *MemRepository) SetCounter(_ context.Context, key string, value int64) error {
	r.storage.Set(key, value)
	return nil
}

func (r *MemRepository) SetGauge(_ context.Context, key string, value float64) error {
	r.storage.Set(key, value)
	return nil
}

func (r *MemRepository) GetAll(_ context.Context) (map[string]any, error) {
	return r.storage.GetAll(), nil
}
