package handlers

import (
	"errors"
	"fmt"
	"go-metrics-service/internal/common/protocol"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type UpdateMetricPathParamsHandler struct {
	metricUpdater MetricUpdater
	logger        *zap.Logger
}

func NewUpdateMetricPathParams(
	metricUpdater MetricUpdater,
	logger *zap.Logger,
) *UpdateMetricPathParamsHandler {
	return &UpdateMetricPathParamsHandler{
		metricUpdater: metricUpdater,
		logger:        logger,
	}
}

func (h *UpdateMetricPathParamsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestLogger := NewRequestLogger(h.logger, r)
	defer closeBody(r.Body, requestLogger)
	metricType := chi.URLParam(r, protocol.TypeParam)
	key := chi.URLParam(r, protocol.KeyParam)
	valueStr := chi.URLParam(r, protocol.ValueParam)
	if len(key) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err := h.updateValue(metricType, key, valueStr, requestLogger)
	if err != nil {
		switch {
		case errors.Is(err, ErrParsing):
			requestLogger.Debug("parsing failed", zap.Error(err))
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

func (h *UpdateMetricPathParamsHandler) updateValue(metricType, key, valueStr string, requestLogger *zap.Logger) error {
	requestLogger.Debug("updating value",
		zap.String("metricType", metricType),
		zap.String("key", key),
		zap.String("value", valueStr),
	)
	switch metricType {
	case protocol.Gauge:
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrParsing, err)
		}
		if err := h.metricUpdater.UpdateOne(protocol.Metrics{
			ID:    key,
			MType: protocol.Gauge,
			Delta: nil,
			Value: &value,
		}); err != nil {
			return fmt.Errorf("failed to set: %w", err)
		}
		return nil
	case protocol.Counter:
		delta, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrParsing, err)
		}
		if err := h.metricUpdater.UpdateOne(protocol.Metrics{
			ID:    key,
			MType: protocol.Gauge,
			Delta: &delta,
			Value: nil,
		}); err != nil {
			return fmt.Errorf("failed to set: %w", err)
		}
		return nil
	default:
		return ErrNonExistentType
	}
}
