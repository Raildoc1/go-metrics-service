package getmetricvalue

import (
	"errors"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data"
	"go-metrics-service/internal/server/logger"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type repository interface {
	SetFloat64(key string, value float64) error
	GetFloat64(key string) (value float64, err error)
	SetInt64(key string, value int64) error
	GetInt64(key string) (value int64, err error)
}

type handler struct {
	repository repository
}

func New(repository repository) http.Handler {
	return &handler{
		repository: repository,
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
	value, err := h.repository.GetFloat64(key)
	if err != nil {
		handleError(w, err)
	} else {
		writeResponse(w, strconv.FormatFloat(value, 'f', -1, 64))
	}
}

func (h *handler) handleCounter(key string, w http.ResponseWriter) {
	value, err := h.repository.GetInt64(key)
	if err != nil {
		handleError(w, err)
	} else {
		writeResponse(w, strconv.FormatInt(value, 10))
	}
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, data.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
		return
	case errors.Is(err, data.ErrWrongType):
		w.WriteHeader(http.StatusBadRequest)
		return
	default:
		logger.Log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func writeResponse(w http.ResponseWriter, value string) {
	w.Header().Set("Content-Type", "text/plain")
	_, err := w.Write([]byte(value))
	if err != nil {
		logger.Log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
