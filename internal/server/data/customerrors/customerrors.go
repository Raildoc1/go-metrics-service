package customerrors

import "errors"

var (
	ErrWrongType = errors.New("wrong type")
	ErrNotFound  = errors.New("not found")
)
