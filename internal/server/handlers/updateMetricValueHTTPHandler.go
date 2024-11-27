package handlers

import (
	"github.com/go-chi/chi/v5"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/logic"
	"net/http"
	"strconv"
)

type updateMetricValueHTTPHandler struct {
	counterLogic *logic.CounterLogic
	gaugeLogic   *logic.GaugeLogic
}

func NewUpdateMetricValueHTTPHandler(counterLogic *logic.CounterLogic, gaugeLogic *logic.GaugeLogic) http.Handler {
	return &updateMetricValueHTTPHandler{
		counterLogic: counterLogic,
		gaugeLogic:   gaugeLogic,
	}
}

func (h *updateMetricValueHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, protocol.KeyParam)
	if len(key) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metricType := chi.URLParam(r, protocol.TypeParam)
	switch metricType {
	case protocol.Gauge:
		h.handleGauge(key, w, r)
	case protocol.Counter:
		h.handleCounter(key, w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (h *updateMetricValueHTTPHandler) handleGauge(key string, w http.ResponseWriter, r *http.Request) {
	valueStr := chi.URLParam(r, protocol.ValueParam)
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.gaugeLogic.Set(key, value); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *updateMetricValueHTTPHandler) handleCounter(key string, w http.ResponseWriter, r *http.Request) {
	deltaStr := chi.URLParam(r, protocol.ValueParam)
	delta, err := strconv.ParseInt(deltaStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.counterLogic.Change(key, delta); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
