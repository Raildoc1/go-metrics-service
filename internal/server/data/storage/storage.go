package storage

import (
	"reflect"
)

type Storage interface {
	Set(key string, value any)
	Get(key string) (any, bool)
	GetAll() map[string]any
}

func Set[T any](s Storage, key string, value T) error {
	if val, ok := s.Get(key); ok {
		if _, ok := val.(T); !ok {
			var zero T
			return WrongTypeError{
				Requested: reflect.TypeOf(zero),
				Actual:    reflect.TypeOf(val),
			}
		}
	}
	s.Set(key, value)
	return nil
}

func Get[T any](s Storage, key string) (T, error) {
	val, ok := s.Get(key)
	if !ok {
		var zero T
		return zero, NotFoundError{
			Key: key,
		}
	}
	castedValue, ok := val.(T)
	if !ok {
		var zero T
		return zero, WrongTypeError{
			Requested: reflect.TypeOf(zero),
			Actual:    reflect.TypeOf(val),
		}
	}
	return castedValue, nil
}
