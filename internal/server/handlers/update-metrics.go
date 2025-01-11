package handlers

import (
	"encoding/json"
	"go-metrics-service/internal/common/protocol"
	"net/http"

	"go.uber.org/zap"
)

type UpdateMetricsValueHandler struct {
	metricUpdater MetricUpdater
	logger        *zap.Logger
}

func NewUpdateMetrics(
	metricUpdater MetricUpdater,
	logger *zap.Logger,
) http.Handler {
	return &UpdateMetricsValueHandler{
		metricUpdater: metricUpdater,
		logger:        logger,
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

	if err := h.metricUpdater.UpdateMany(requestData); err != nil {
		h.logger.Error("failed to update metrics", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
