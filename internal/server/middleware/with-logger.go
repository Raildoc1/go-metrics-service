package middleware

import (
	"go-metrics-service/internal/server/logger"
	"net/http"
	"time"
)

func WithLogger(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := loggingResponseWriter{
			inner: w,
		}
		inner.ServeHTTP(&lrw, r)
		logger.Log.Infoln(
			"method", r.Method,
			"url", r.URL.String(),
			"duration", time.Since(start),
			"status", lrw.status,
			"size", lrw.size,
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

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.inner.Write(b)
	w.size += size
	return size, err
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.inner.WriteHeader(status)
	w.status = status
}
