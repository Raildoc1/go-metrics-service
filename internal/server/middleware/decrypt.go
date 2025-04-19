package middleware

import (
	"bytes"
	"go-metrics-service/internal/server/handlers"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type Decoder interface {
	Decode([]byte) ([]byte, error)
}

type RequestDecoder struct {
	decoder Decoder
	logger  *zap.Logger
}

func NewRequestDecoder(decoder Decoder, logger *zap.Logger) *RequestDecoder {
	return &RequestDecoder{
		decoder: decoder,
		logger:  logger,
	}
}

func (rd *RequestDecoder) CreateHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger := handlers.NewRequestLogger(rd.logger, r)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			requestLogger.Error("failed to read body ", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		decoded, err := rd.decoder.Decode(body)
		if err != nil {
			requestLogger.Error("failed to decode body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(decoded))

		next.ServeHTTP(w, r)
	})
}
