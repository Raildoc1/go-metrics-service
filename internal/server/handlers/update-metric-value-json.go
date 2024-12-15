package handlers

import (
	"encoding/json"
	"go-metrics-service/internal/common/protocol"
	"io"
	"net/http"
)

type UpdateMetricValueJsonHandler struct {
	counterLogic CounterLogic
	gaugeLogic   GaugeLogic
	logger       Logger
}

func NewUpdateMetricValueJsonHandler(
	counterLogic CounterLogic,
	gaugeLogic GaugeLogic,
	logger Logger,
) http.Handler {
	return &UpdateMetricValueJsonHandler{
		counterLogic: counterLogic,
		gaugeLogic:   gaugeLogic,
		logger:       logger,
	}
}

func (h *UpdateMetricValueJsonHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	switch requestData.MType {
	case protocol.Gauge:
		if err := h.gaugeLogic.Set(requestData.ID, *requestData.Value); err != nil {
			h.logger.Errorln(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case protocol.Counter:
		if err := h.counterLogic.Change(requestData.ID, *requestData.Delta); err != nil {
			h.logger.Errorln(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		h.logger.Debugln("POST /update wrong metric type: ", requestData.MType)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
