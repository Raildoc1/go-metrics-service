package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data"
	"io"
	"net/http"
)

type UpdateMetricValueHandler struct {
	counterLogic CounterLogic
	gaugeLogic   GaugeLogic
	logger       Logger
}

func NewUpdateMetric(
	counterLogic CounterLogic,
	gaugeLogic GaugeLogic,
	logger Logger,
) http.Handler {
	return &UpdateMetricValueHandler{
		counterLogic: counterLogic,
		gaugeLogic:   gaugeLogic,
		logger:       logger,
	}
}

func (h *UpdateMetricValueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			h.logger.Errorln(err)
		}
	}(r.Body)

	var requestData protocol.Metrics
	jsonDecoder := json.NewDecoder(r.Body)
	if err := jsonDecoder.Decode(&requestData); err != nil {
		h.logger.Debugln(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.update(&requestData); err != nil {
		switch {
		case errors.Is(err, ErrWrongValueType):
			h.logger.Debugln(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		case errors.Is(err, data.ErrWrongType):
			h.logger.Debugln(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		case errors.Is(err, ErrNonExistentType):
			h.logger.Debugln(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		default:
			h.logger.Errorln(err)
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
		return fmt.Errorf("%w:  %s ", ErrNonExistentType, requestData.MType)
	}
	return nil
}
