package handlers

import (
	"errors"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type GetMetricValueTextHandler struct {
	gaugeRepository   GaugeRepository
	counterRepository CounterRepository
	logger            Logger
}

func NewGetMetricValueTextHandler(
	gaugeRepository GaugeRepository,
	counterRepository CounterRepository,
	logger Logger,
) *GetMetricValueTextHandler {
	return &GetMetricValueTextHandler{
		gaugeRepository:   gaugeRepository,
		counterRepository: counterRepository,
		logger:            logger,
	}
}

func (h *GetMetricValueTextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, protocol.KeyParam)
	if len(key) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metricType := chi.URLParam(r, protocol.TypeParam)
	switch metricType {
	case protocol.Gauge:
		h.handleGauge(key, w)
	case protocol.Counter:
		h.handleCounter(key, w)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (h *GetMetricValueTextHandler) handleGauge(key string, w http.ResponseWriter) {
	value, err := h.gaugeRepository.GetFloat64(key)
	if err != nil {
		h.handleError(w, err)
	} else {
		h.writeResponse(w, strconv.FormatFloat(value, 'f', -1, 64))
	}
}

func (h *GetMetricValueTextHandler) handleCounter(key string, w http.ResponseWriter) {
	value, err := h.counterRepository.GetInt64(key)
	if err != nil {
		h.handleError(w, err)
	} else {
		h.writeResponse(w, strconv.FormatInt(value, 10))
	}
}

func (h *GetMetricValueTextHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, data.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
		return
	case errors.Is(err, data.ErrWrongType):
		w.WriteHeader(http.StatusBadRequest)
		return
	default:
		h.logger.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *GetMetricValueTextHandler) writeResponse(w http.ResponseWriter, value string) {
	w.Header().Set("Content-Type", "text/plain")
	_, err := w.Write([]byte(value))
	if err != nil {
		h.logger.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
