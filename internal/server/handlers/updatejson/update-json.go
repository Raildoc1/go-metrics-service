package updatejson

import (
	"encoding/json"
	"go-metrics-service/internal/common/protocol"
	"io"
	"net/http"
)

type CounterLogic interface {
	Change(key string, delta int64) error
}

type GaugeLogic interface {
	Set(key string, value float64) error
}

type Logger interface {
	Error(args ...interface{})
	Debug(args ...interface{})
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			h.logger.Error(err)
		}
	}(r.Body)

	var requestData protocol.Metrics
	jsonDecoder := json.NewDecoder(r.Body)
	if err := jsonDecoder.Decode(&requestData); err != nil {
		h.logger.Debug(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch requestData.MType {
	case protocol.Gauge:
		if err := h.gaugeLogic.Set(requestData.ID, *requestData.Value); err != nil {
			h.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case protocol.Counter:
		if err := h.counterLogic.Change(requestData.ID, *requestData.Delta); err != nil {
			h.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		h.logger.Debug("POST /update wrong metric type: ", requestData.MType)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
