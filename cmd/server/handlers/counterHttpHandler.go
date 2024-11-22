package handlers

import (
	"go-metrics-service/cmd/server/logic"
	"net/http"
	"strconv"
	"strings"
)

type CounterHttpHandler struct {
	counterLogic *logic.CounterLogic
}

func NewCounterHttpHandler(counterLogic *logic.CounterLogic) *CounterHttpHandler {
	return &CounterHttpHandler{counterLogic}
}

func (chh *CounterHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	key := pathParts[len(pathParts)-2]
	valueStr := pathParts[len(pathParts)-1]
	delta, err := strconv.ParseInt(valueStr, 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	chh.counterLogic.Change(key, delta)
}
