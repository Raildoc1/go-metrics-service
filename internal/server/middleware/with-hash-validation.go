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

func withHashValidation(h http.Handler, hasher hash.Hash, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hasher == nil {
			h.ServeHTTP(w, r)
			return
		}
		requestLogger := handlers.NewRequestLogger(logger, r)
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
		_, err = hasher.Write(bodyBytes)
		if err != nil {
			requestLogger.Error("failed to write body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
		calculatedHashVal := hex.EncodeToString(hasher.Sum(nil))
		receivedHashVal := r.Header.Get(protocol.HashHeader)
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
		h.ServeHTTP(w, r)
	})
}
