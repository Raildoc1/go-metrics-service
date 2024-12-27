package handlers

import (
	"net/http"

	"go.uber.org/zap"
)

func NewRequestLogger(logger *zap.Logger, r *http.Request) *zap.Logger {
	return logger.With(
		zap.String("path", r.URL.Path),
		zap.String("method", r.Method),
		zap.String("remote-addr", r.RemoteAddr),
	)
}
