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

func withHash(h http.Handler, hh hash.Hash, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hh == nil {
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
		_, err = hh.Write(bodyBytes)
		if err != nil {
			requestLogger.Error("failed to write body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
		calculatedHashVal := hex.EncodeToString(hh.Sum(nil))
		receivedHashVal := r.Header.Get(protocol.HashHeader)
		if calculatedHashVal != receivedHashVal {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		h.ServeHTTP(w, r)
	})
}
