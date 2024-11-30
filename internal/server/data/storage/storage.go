package storage

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	WrongTypeError = errors.New("wrong type")
	NotFoundError  = errors.New("not found")
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
			return CreateWrongTypeError(reflect.TypeOf(zero), reflect.TypeOf(val))
		}
	}
	s.Set(key, value)
	return nil
}

func Get[T any](s Storage, key string) (T, error) {
	val, ok := s.Get(key)
	if !ok {
		var zero T
		return zero, CreateNotFoundError(key)
	}
	castedValue, ok := val.(T)
	if !ok {
		var zero T
		return zero, CreateWrongTypeError(reflect.TypeOf(zero), reflect.TypeOf(val))
	}
	return castedValue, nil
}

func CreateNotFoundError(key string) error {
	return fmt.Errorf(
		"%w: '%s' not found",
		NotFoundError,
		key,
	)
}

func CreateWrongTypeError(requested, actual reflect.Type) error {
	return fmt.Errorf(
		"%w: expected type %s but data contains %s",
		WrongTypeError,
		requested,
		actual,
	)
}
