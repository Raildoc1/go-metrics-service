package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func withLogger(inner http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqLogger := logger.With(
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		start := time.Now()
		lrw := loggingResponseWriter{
			inner:  w,
			status: http.StatusOK,
			size:   0,
		}
		inner.ServeHTTP(&lrw, r)
		reqLogger.Info(
			"Request handled",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
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
