package middleware

import (
	"bytes"
	"encoding/hex"
	"go-metrics-service/internal/common/protocol"
	"hash"
	"net/http"

	"go.uber.org/zap"
)

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

func withResponseHashing(h http.Handler, hasher hash.Hash, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hasher == nil {
			h.ServeHTTP(w, r)
			return
		}
		hw := &HashWriter{
			ResponseWriter: w,
			h:              hasher,
		}
		h.ServeHTTP(hw, r)
		w.Header().Set(protocol.HashHeader, hex.EncodeToString(hw.h.Sum(nil)))
		_, err := hw.b.WriteTo(w)
		if err != nil {
			logger.Error("failed to write response", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
