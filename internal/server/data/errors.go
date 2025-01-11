package data

import "errors"

var (
	ErrWrongType           = errors.New("wrong type")
	ErrNotFound            = errors.New("not found")
	ErrWrongTransactionID  = errors.New("wrong transaction id")
	ErrNoTransactionOpened = errors.New("no transaction opened")
)
