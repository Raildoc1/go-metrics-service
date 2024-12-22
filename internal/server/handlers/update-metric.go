package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data"
	"net/http"

	"go.uber.org/zap"
)

type UpdateMetricValueHandler struct {
	counterLogic CounterLogic
	gaugeLogic   GaugeLogic
	logger       *zap.Logger
}

func NewUpdateMetric(
	counterLogic CounterLogic,
	gaugeLogic GaugeLogic,
	logger *zap.Logger,
) http.Handler {
	return &UpdateMetricValueHandler{
		counterLogic: counterLogic,
		gaugeLogic:   gaugeLogic,
		logger:       logger,
	}
}

func (h *UpdateMetricValueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestLogger := NewRequestLogger(h.logger, r)
	defer closeBody(r.Body, requestLogger)

	var requestData protocol.Metrics
	jsonDecoder := json.NewDecoder(r.Body)
	if err := jsonDecoder.Decode(&requestData); err != nil {
		requestLogger.Debug("failed to decode request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	const errUpdate = "update failed"
	if err := h.update(&requestData); err != nil {
		switch {
		case errors.Is(err, ErrWrongValueType):
			requestLogger.Debug(errUpdate, zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		case errors.Is(err, data.ErrWrongType):
			requestLogger.Debug(errUpdate, zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		case errors.Is(err, ErrNonExistentType):
			requestLogger.Debug(errUpdate, zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		default:
			requestLogger.Error(errUpdate, zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (h *UpdateMetricValueHandler) update(requestData *protocol.Metrics) error {
	switch requestData.MType {
	case protocol.Gauge:
		if requestData.Value == nil {
			return ErrWrongValueType
		}
		if err := h.gaugeLogic.Set(requestData.ID, *requestData.Value); err != nil {
			return fmt.Errorf("set gauge: %w", err)
		}
	case protocol.Counter:
		if requestData.Delta == nil {
			return ErrWrongValueType
		}
		if err := h.counterLogic.Change(requestData.ID, *requestData.Delta); err != nil {
			return fmt.Errorf("change counter: %w", err)
		}
	default:
		return ErrNonExistentType
	}
	return nil
}
