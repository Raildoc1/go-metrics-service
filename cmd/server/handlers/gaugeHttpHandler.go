package handlers

import (
	"go-metrics-service/cmd/server/logic"
	"net/http"
	"strconv"
	"strings"
)

type GaugeHttpHandler struct {
	gaugeLogic *logic.GaugeLogic
}

func NewGaugeHttpHandler(gaugeLogic *logic.GaugeLogic) *GaugeHttpHandler {
	return &GaugeHttpHandler{gaugeLogic}
}

func (ghh *GaugeHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	key := pathParts[len(pathParts)-2]
	valueStr := pathParts[len(pathParts)-1]
	value, err := strconv.ParseFloat(valueStr, 64)

	if len(key) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ghh.gaugeLogic.Set(key, value)
}
