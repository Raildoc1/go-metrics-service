package handlers

import (
	"errors"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data/repositories"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type getMetricValueHTTPHandler struct {
	counterRepository repositories.Repository[int64]
	gaugeRepository   repositories.Repository[float64]
}

func NewGetMetricValueHTTPHandler(
	counterRepository repositories.Repository[int64],
	gaugeRepository repositories.Repository[float64],
) http.Handler {
	return &getMetricValueHTTPHandler{
		counterRepository: counterRepository,
		gaugeRepository:   gaugeRepository,
	}
}

func (h *getMetricValueHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h *getMetricValueHTTPHandler) handleGauge(key string, w http.ResponseWriter) {
	value, err := h.gaugeRepository.Get(key)
	if err != nil {
		handleError(w, err)
	} else {
		writeResponse(w, strconv.FormatFloat(value, 'f', -1, 64))
	}
}

func (h *getMetricValueHTTPHandler) handleCounter(key string, w http.ResponseWriter) {
	value, err := h.counterRepository.Get(key)
	if err != nil {
		handleError(w, err)
	} else {
		writeResponse(w, strconv.FormatInt(value, 10))
	}
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repositories.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
		return
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func writeResponse(w http.ResponseWriter, value string) {
	w.Header().Set("Content-Type", "text/plain")
	_, err := w.Write([]byte(value))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
