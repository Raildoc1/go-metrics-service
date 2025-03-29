package middleware

import (
	"bytes"
	"encoding/hex"
	"fmt"
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
		bodyBytes, err := readBodyAndRewind(&r.Body)
		if err != nil {
			requestLogger.Error("failed to read request body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		calculatedHashVal, err := calculateHash(bodyBytes, rh.hashFactory)
		if err != nil {
			requestLogger.Error("failed to calculate hash", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
		if calculatedHashVal != receivedHashVal {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func calculateHash(bodyBytes []byte, hashFactory HashFactory) (string, error) {
	h := hashFactory.Create()
	_, err := h.Write(bodyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to write body: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func readBodyAndRewind(readCloser *io.ReadCloser) ([]byte, error) {
	result := bytes.Buffer{}
	_, err := io.Copy(&result, *readCloser)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	err = (*readCloser).Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close body: %w", err)
	}
	*readCloser = io.NopCloser(bytes.NewBuffer(result.Bytes()))
	return result.Bytes(), nil
}
