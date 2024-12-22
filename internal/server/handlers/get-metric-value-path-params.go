package handlers

import (
	"errors"
	"fmt"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type GetMetricValuePathParamsHandler struct {
	gaugeRepository   GaugeRepository
	counterRepository CounterRepository
	logger            *zap.Logger
}

func NewGetMetricValuePathParams(
	gaugeRepository GaugeRepository,
	counterRepository CounterRepository,
	logger *zap.Logger,
) *GetMetricValuePathParamsHandler {
	return &GetMetricValuePathParamsHandler{
		gaugeRepository:   gaugeRepository,
		counterRepository: counterRepository,
		logger:            logger,
	}
}

func (h *GetMetricValuePathParamsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestLogger := NewRequestLogger(h.logger, r)
	defer closeBody(r.Body, requestLogger)
	metricType := chi.URLParam(r, protocol.TypeParam)
	key := chi.URLParam(r, protocol.KeyParam)
	if len(key) == 0 {
		requestLogger.Debug("empty key requested")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	result, err := h.getValue(metricType, key)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNotFound):
			requestLogger.Debug("metric not found", zap.Error(err))
			w.WriteHeader(http.StatusNotFound)
			return
		case errors.Is(err, data.ErrWrongType):
			requestLogger.Debug("wrong type", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		default:
			requestLogger.Error("unexpected error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "text/plain")
	_, err = w.Write([]byte(result))
	if err != nil {
		requestLogger.Error("failed to write response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *GetMetricValuePathParamsHandler) getValue(metricType, key string) (string, error) {
	switch metricType {
	case protocol.Gauge:
		value, err := h.gaugeRepository.GetFloat64(key)
		if err != nil {
			return "", fmt.Errorf("get gauge: %w", err)
		}
		return strconv.FormatFloat(value, 'f', -1, 64), nil
	case protocol.Counter:
		value, err := h.counterRepository.GetInt64(key)
		if err != nil {
			return "", fmt.Errorf("get counter: %w", err)
		}
		return strconv.FormatInt(value, 10), nil
	default:
		return "", fmt.Errorf("%w:  %s ", ErrNonExistentType, metricType)
	}
}
