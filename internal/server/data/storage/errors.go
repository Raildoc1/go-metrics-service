package storage

import (
	"fmt"
	"reflect"
)

type WrongTypeError struct {
	Requested reflect.Type
	Actual    reflect.Type
}

func (e WrongTypeError) Error() string {
	return fmt.Sprintf("expected type %s but data contains %s", e.Requested, e.Actual)
}

type NotFoundError struct {
	Key string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("key %s not found", e.Key)
}
