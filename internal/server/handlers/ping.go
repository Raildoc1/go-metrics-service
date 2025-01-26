package handlers

import (
	"net/http"

	"go.uber.org/zap"
)

type Pingable interface {
	Ping() error
}

type PingHandler struct {
	logger    *zap.Logger
	pingables []Pingable
}

func NewPing(
	pingables []Pingable,
	logger *zap.Logger,
) http.Handler {
	return &PingHandler{
		pingables: pingables,
		logger:    logger,
	}
}

func (h *PingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, pingable := range h.pingables {
		err := pingable.Ping()
		if err != nil {
			h.logger.Error("database ping error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
