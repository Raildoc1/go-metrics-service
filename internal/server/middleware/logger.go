package middleware

import (
	"go-metrics-service/internal/server/handlers"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Logger struct {
	logger *zap.Logger
}

func NewLogger(logger *zap.Logger) *Logger {
	return &Logger{
		logger: logger,
	}
}

func (l *Logger) CreateHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger := handlers.NewRequestLogger(l.logger, r)
		start := time.Now()
		lrw := loggingResponseWriter{
			inner:  w,
			status: http.StatusOK,
			size:   0,
		}
		next.ServeHTTP(&lrw, r)
		requestLogger.Info(
			"Request handled",
			zap.Duration("duration", time.Since(start)),
			zap.Int("status", lrw.status),
			zap.Int("size", lrw.size),
		)
	})
}

type loggingResponseWriter struct {
	inner  http.ResponseWriter
	status int
	size   int
}

func (w *loggingResponseWriter) Header() http.Header {
	return w.inner.Header()
}

//nolint:wrapcheck // logging middleware must not change inner error
func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.inner.Write(b)
	w.size += size
	return size, err
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.inner.WriteHeader(status)
	w.status = status
}
