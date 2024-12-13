package updatemetricvalue

import (
	"go-metrics-service/internal/common/protocol"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CounterLogic interface {
	Change(key string, delta int64) error
}

type GaugeLogic interface {
	Set(key string, value float64) error
}

type Logger interface {
	Error(args ...interface{})
}

type handler struct {
	counterLogic CounterLogic
	gaugeLogic   GaugeLogic
	logger       Logger
}

func New(
	counterLogic CounterLogic,
	gaugeLogic GaugeLogic,
	logger Logger,
) http.Handler {
	return &handler{
		counterLogic: counterLogic,
		gaugeLogic:   gaugeLogic,
		logger:       logger,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) handleGauge(key string, w http.ResponseWriter, r *http.Request) {
	valueStr := chi.URLParam(r, protocol.ValueParam)
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.gaugeLogic.Set(key, value); err != nil {
		h.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *handler) handleCounter(key string, w http.ResponseWriter, r *http.Request) {
	deltaStr := chi.URLParam(r, protocol.ValueParam)
	delta, err := strconv.ParseInt(deltaStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.counterLogic.Change(key, delta); err != nil {
		h.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
