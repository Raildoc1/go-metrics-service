package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type compressReader struct {
	originalReader      io.ReadCloser
	decompressingReader *gzip.Reader
}

func newCompressReader(reader io.ReadCloser) (*compressReader, error) {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	return &compressReader{
		originalReader:      reader,
		decompressingReader: gzipReader,
	}, nil
}

//nolint:wrapcheck // wrapping unnecessary
func (r *compressReader) Read(p []byte) (n int, err error) {
	return r.decompressingReader.Read(p)
}

//nolint:wrapcheck // wrapping unnecessary
func (r *compressReader) Close() error {
	if err := r.originalReader.Close(); err != nil {
		return err
	}
	return r.decompressingReader.Close()
}

func withRequestDecompression(h http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqLogger := logger.With(
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)

		if r.Header.Get("Content-Encoding") != "gzip" {
			h.ServeHTTP(w, r)
			return
		}

		decompressingReader, err := newCompressReader(r.Body)

		if err != nil {
			reqLogger.Error("failed to decompress ", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = decompressingReader

		h.ServeHTTP(w, r)
	})
}
