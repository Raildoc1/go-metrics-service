package server

import (
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data/repository"
	"go-metrics-service/internal/server/handlers/getallmetrics"
	"go-metrics-service/internal/server/handlers/getmetricvalue"
	"go-metrics-service/internal/server/handlers/updatemetricvalue"
	"go-metrics-service/internal/server/logic/counter"
	"go-metrics-service/internal/server/logic/gauge"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewServer(storage repository.Storage) http.Handler {
	rep := repository.New(storage)

	counterLogic := counter.New(rep)
	gaugeLogic := gauge.New(rep)

	updateMetricValueHTTPHandler := updatemetricvalue.New(
		counterLogic,
		gaugeLogic,
	)

	getMetricValueHTTPHandler := getmetricvalue.New(rep)
	getAllMetricsHTTPHandler := getallmetrics.New(storage)

	router := chi.NewRouter()

	router.Post(protocol.UpdateMetricValueURL, updateMetricValueHTTPHandler.ServeHTTP)
	router.Get(protocol.GetMetricValueURL, getMetricValueHTTPHandler.ServeHTTP)
	router.Get(protocol.GetAllMetricsURL, getAllMetricsHTTPHandler.ServeHTTP)

	return router
}
