package repositories

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrWrongType = errors.New("wrong type")
	ErrNotFound  = errors.New("not found")
)

type Storage interface {
	Set(key string, value any)
	Has(key string) bool
	Get(key string) (any, bool)
	GetAll() map[string]any
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
		ErrNotFound,
		key,
	)
}

func createWrongTypeError(requested, actual reflect.Type) error {
	return fmt.Errorf(
		"%w: expected type %s but data contains %s",
		ErrWrongType,
		requested,
		actual,
	)
}
