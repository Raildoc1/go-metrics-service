package middleware

import (
	"go-metrics-service/internal/server/handlers"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func withLogger(inner http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger := handlers.NewRequestLogger(logger, r)
		start := time.Now()
		lrw := loggingResponseWriter{
			inner:  w,
			status: http.StatusOK,
			size:   0,
		}
		inner.ServeHTTP(&lrw, r)
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
