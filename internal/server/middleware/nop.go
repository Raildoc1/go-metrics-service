package middleware

import (
	"net/http"
)

type Nop struct{}

func NewNop() *Nop {
	return &Nop{}
}

func (i *Nop) CreateHandler(next http.Handler) http.Handler {
	return next
}
