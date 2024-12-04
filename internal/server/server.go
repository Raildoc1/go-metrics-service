package server

import (
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data/repositories"
	"go-metrics-service/internal/server/handlers/getallmetrics"
	"go-metrics-service/internal/server/handlers/getmetricvalue"
	"go-metrics-service/internal/server/handlers/updatemetricvalue"
	"go-metrics-service/internal/server/logic/counter"
	"go-metrics-service/internal/server/logic/gauge"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewServer(storage repositories.Storage) http.Handler {
	counterRepository := repositories.NewCounterRepository(storage)
	gaugeRepository := repositories.NewGaugeRepository(storage)

	counterLogic := counter.New(counterRepository)
	gaugeLogic := gauge.New(gaugeRepository)

	updateMetricValueHTTPHandler := updatemetricvalue.New(
		counterLogic,
		gaugeLogic,
	)

	getMetricValueHTTPHandler := getmetricvalue.New(
		counterRepository,
		gaugeRepository,
	)

	getAllMetricsHTTPHandler := getallmetrics.New(
		storage,
	)

	router := chi.NewRouter()

	router.Post(protocol.UpdateMetricValueURL, updateMetricValueHTTPHandler.ServeHTTP)
	router.Get(protocol.GetMetricValueURL, getMetricValueHTTPHandler.ServeHTTP)
	router.Get(protocol.GetAllMetricsURL, getAllMetricsHTTPHandler.ServeHTTP)

	return router
}
