package repository

import (
	"fmt"
	"go-metrics-service/internal/server/data"
	"reflect"
)

type Storage interface {
	Set(key string, value any)
	Has(key string) bool
	Get(key string) (any, bool)
	GetAll() map[string]any
}

type Repository struct {
	storage Storage
}

func New(storage Storage) *Repository {
	return &Repository{
		storage: storage,
	}
}

func (r *Repository) Has(key string) bool {
	return r.storage.Has(key)
}

func (r *Repository) SetInt64(key string, value int64) error {
	return set[int64](r.storage, key, value)
}

func (r *Repository) GetInt64(key string) (int64, error) {
	return get[int64](r.storage, key)
}

func (r *Repository) SetFloat64(key string, value float64) error {
	return set[float64](r.storage, key, value)
}

func (r *Repository) GetFloat64(key string) (float64, error) {
	return get[float64](r.storage, key)
}

func set[T any](s Storage, key string, value T) error {
	if val, ok := s.Get(key); ok {
		if _, ok := val.(T); !ok {
			var zero T
			return createWrongTypeError(reflect.TypeOf(zero), reflect.TypeOf(val))
		}
	}
	s.Set(key, value)
	return nil
}

func get[T any](s Storage, key string) (T, error) {
	val, ok := s.Get(key)
	if !ok {
		var zero T
		return zero, createNotFoundError(key)
	}
	castedValue, ok := val.(T)
	if !ok {
		var zero T
		return zero, createWrongTypeError(reflect.TypeOf(zero), reflect.TypeOf(val))
	}
	return castedValue, nil
}

func createNotFoundError(key string) error {
	return fmt.Errorf(
		"%w: '%s' not found",
		data.ErrNotFound,
		key,
	)
}

func createWrongTypeError(requested, actual reflect.Type) error {
	return fmt.Errorf(
		"%w: expected type %s but data contains %s",
		data.ErrWrongType,
		requested,
		actual,
	)
}
