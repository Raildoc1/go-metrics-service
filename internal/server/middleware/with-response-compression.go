package middleware

import (
	"compress/gzip"
	"go-metrics-service/internal/server/handlers"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer           io.Writer
	uncompressedSize int
}

//nolint:wrapcheck // wrapping unnecessary
func (w *gzipWriter) Write(b []byte) (int, error) {
	w.uncompressedSize += len(b)
	return w.Writer.Write(b)
}

func withResponseCompression(h http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger := handlers.NewRequestLogger(logger, r)

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			requestLogger.Debug("compression missed")
			h.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			requestLogger.Error("Failed to create gzip writer", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func(gz *gzip.Writer) {
			err := gz.Close()
			if err != nil {
				requestLogger.Error("Failed to close gzip writer", zap.Error(err))
				return
			}
		}(gz)

		w.Header().Set("Content-Encoding", "gzip")
		wrappedWriter := gzipWriter{ResponseWriter: w, Writer: gz}
		h.ServeHTTP(&wrappedWriter, r)

		requestLogger.Debug("request compressed",
			zap.Int("Uncompressed size", wrappedWriter.uncompressedSize),
		)
	})
}
