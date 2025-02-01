package middleware

import (
	"bytes"
	"encoding/hex"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/handlers"
	"hash"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type HashFactory interface {
	Create() hash.Hash
}

type ResponseHash struct {
	hashFactory HashFactory
	logger      *zap.Logger
}

type HashWriter struct {
	http.ResponseWriter
	h hash.Hash
	b bytes.Buffer
}

//nolint:wrapcheck // wrapping unnecessary
func (w *HashWriter) Write(b []byte) (int, error) {
	w.h.Write(b)
	return w.b.Write(b)
}

func NewResponseHash(logger *zap.Logger, hashFactory HashFactory) *ResponseHash {
	return &ResponseHash{
		hashFactory: hashFactory,
		logger:      logger,
	}
}

func (rh *ResponseHash) CreateHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hw := &HashWriter{
			ResponseWriter: w,
			h:              rh.hashFactory.Create(),
		}
		h.ServeHTTP(hw, r)
		w.Header().Set(protocol.HashHeader, hex.EncodeToString(hw.h.Sum(nil)))
		_, err := hw.b.WriteTo(w)
		if err != nil {
			rh.logger.Error("failed to write response", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

type RequestHash struct {
	hashFactory HashFactory
	logger      *zap.Logger
}

func NewRequestHash(logger *zap.Logger, hashFactory HashFactory) *RequestHash {
	return &RequestHash{
		hashFactory: hashFactory,
		logger:      logger,
	}
}

func (rh *RequestHash) CreateHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHashVal := r.Header.Get(protocol.HashHeader)
		if receivedHashVal == "" {
			next.ServeHTTP(w, r)
			return
		}
		requestLogger := handlers.NewRequestLogger(rh.logger, r)
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			requestLogger.Error("failed to read body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h := rh.hashFactory.Create()
		_, err = h.Write(bodyBytes)
		if err != nil {
			requestLogger.Error("failed to write body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
		calculatedHashVal := hex.EncodeToString(h.Sum(nil))
		if calculatedHashVal != receivedHashVal {
			requestLogger.Debug(
				"hash mismatch",
				zap.String("calculatedHash", calculatedHashVal),
				zap.String("receivedHash", receivedHashVal),
			)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		next.ServeHTTP(w, r)
	})
}
