package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data"
	"net/http"

	"go.uber.org/zap"
)

type GetMetricValueHandler struct {
	gaugeRepository   GaugeRepository
	counterRepository CounterRepository
	logger            *zap.Logger
}

func NewGetMetricValue(
	gaugeRepository GaugeRepository,
	counterRepository CounterRepository,
	logger *zap.Logger,
) *GetMetricValueHandler {
	return &GetMetricValueHandler{
		gaugeRepository:   gaugeRepository,
		counterRepository: counterRepository,
		logger:            logger,
	}
}

func (h *GetMetricValueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestLogger := NewRequestLogger(h.logger, r)
	defer closeBody(r.Body, requestLogger)

	var requestData protocol.Metrics

	jsonDecoder := json.NewDecoder(r.Body)
	if err := jsonDecoder.Decode(&requestData); err != nil {
		requestLogger.Debug("failed to decode request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	const errFill = "failed to fill request data"
	if err := h.fill(r.Context(), &requestData); err != nil {
		switch {
		case errors.Is(err, data.ErrNotFound):
			requestLogger.Debug(errFill, zap.Error(err))
			w.WriteHeader(http.StatusNotFound)
			return
		case errors.Is(err, data.ErrWrongType):
			requestLogger.Debug(errFill, zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		case errors.Is(err, ErrNonExistentType):
			requestLogger.Debug(errFill, zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		default:
			requestLogger.Error(errFill, zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	encoded, err := json.Marshal(requestData)
	if err != nil {
		requestLogger.Error("failed to marsha json", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(encoded)
	if err != nil {
		requestLogger.Error("failed to write response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *GetMetricValueHandler) fill(ctx context.Context, requestData *protocol.Metrics) error {
	switch requestData.MType {
	case protocol.Gauge:
		value, err := h.gaugeRepository.GetGauge(ctx, requestData.ID)
		if err != nil {
			return fmt.Errorf("get gauge: %w", err)
		}
		requestData.Value = &value
	case protocol.Counter:
		value, err := h.counterRepository.GetCounter(ctx, requestData.ID)
		if err != nil {
			return fmt.Errorf("get counter: %w", err)
		}
		requestData.Delta = &value
	default:
		return ErrNonExistentType
	}
	return nil
}
