package handlers

import (
	"net/http"

	"go.uber.org/zap"
)

type Database interface {
	Ping() error
}

type PingHandler struct {
	db     Database
	logger *zap.Logger
}

func NewPing(
	db Database,
	logger *zap.Logger,
) http.Handler {
	return &PingHandler{
		db:     db,
		logger: logger,
	}
}

func (h *PingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.db.Ping()
	if err != nil {
		h.logger.Error("database ping error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
