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

type GetMetricValueJsonHandler struct {
	gaugeRepository   GaugeRepository
	counterRepository CounterRepository
	logger            Logger
}

func NewGetMetricValueJsonHandler(
	gaugeRepository GaugeRepository,
	counterRepository CounterRepository,
	logger Logger,
) *GetMetricValueJsonHandler {
	return &GetMetricValueJsonHandler{
		gaugeRepository:   gaugeRepository,
		counterRepository: counterRepository,
		logger:            logger,
	}
}

func (h *GetMetricValueJsonHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	if err := h.fill(&requestData); err != nil {
		switch {
		case errors.Is(err, data.ErrNotFound):
			h.logger.Debugln(err)
			w.WriteHeader(http.StatusNotFound)
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

	encoded, err := json.Marshal(requestData)
	if err != nil {
		h.logger.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(encoded)
	if err != nil {
		h.logger.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *GetMetricValueJsonHandler) fill(requestData *protocol.Metrics) error {
	switch requestData.MType {
	case protocol.Gauge:
		value, err := h.gaugeRepository.GetFloat64(requestData.ID)
		if err != nil {
			return err
		}
		requestData.Value = &value
	case protocol.Counter:
		value, err := h.counterRepository.GetInt64(requestData.ID)
		if err != nil {
			return err
		}
		requestData.Delta = &value
	default:
		return fmt.Errorf("%w:  %s ", ErrNonExistentType, requestData.MType)
	}
	return nil
}
