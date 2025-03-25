package handlers

import (
	"encoding/json"
	"errors"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data"
	"net/http"

	"go.uber.org/zap"
)

type UpdateMetricsValueHandler struct {
	metricController MetricController
	logger           *zap.Logger
}

func NewUpdateMetrics(
	metricController MetricController,
	logger *zap.Logger,
) http.Handler {
	return &UpdateMetricsValueHandler{
		metricController: metricController,
		logger:           logger,
	}
}

func (h *UpdateMetricsValueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestLogger := NewRequestLogger(h.logger, r)
	defer closeBody(r.Body, requestLogger)

	var requestData []protocol.Metrics
	jsonDecoder := json.NewDecoder(r.Body)
	if err := jsonDecoder.Decode(&requestData); err != nil {
		requestLogger.Debug("failed to decode request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.metricController.UpdateMany(r.Context(), requestData); err != nil {
		switch {
		case errors.Is(err, ErrParsing):
			requestLogger.Debug("parsing failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		case errors.Is(err, data.ErrWrongType):
			requestLogger.Debug("wrong type", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		case errors.Is(err, ErrNonExistentType):
			requestLogger.Debug("non-existent type", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		default:
			requestLogger.Error("unexpected error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
