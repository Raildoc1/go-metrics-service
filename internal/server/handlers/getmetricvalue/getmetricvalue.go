package getmetricvalue

import (
	"errors"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data/repositories"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type gaugeRepository interface {
	Set(key string, value float64) error
	Get(key string) (value float64, err error)
}

type counterRepository interface {
	Set(key string, value int64) error
	Get(key string) (value int64, err error)
}

type handler struct {
	counterRepository counterRepository
	gaugeRepository   gaugeRepository
}

func New(
	counterRepository counterRepository,
	gaugeRepository gaugeRepository,
) http.Handler {
	return &handler{
		counterRepository: counterRepository,
		gaugeRepository:   gaugeRepository,
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
		h.handleGauge(key, w)
	case protocol.Counter:
		h.handleCounter(key, w)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (h *handler) handleGauge(key string, w http.ResponseWriter) {
	value, err := h.gaugeRepository.Get(key)
	if err != nil {
		handleError(w, err)
	} else {
		writeResponse(w, strconv.FormatFloat(value, 'f', -1, 64))
	}
}

func (h *handler) handleCounter(key string, w http.ResponseWriter) {
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
	case errors.Is(err, repositories.ErrWrongType):
		w.WriteHeader(http.StatusBadRequest)
		return
	default:
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func writeResponse(w http.ResponseWriter, value string) {
	w.Header().Set("Content-Type", "text/plain")
	_, err := w.Write([]byte(value))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
