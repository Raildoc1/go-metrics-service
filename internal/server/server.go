package server

import (
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data/repositories"
	"go-metrics-service/internal/server/handlers"
	"go-metrics-service/internal/server/logic"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewServer(storage repositories.Storage) http.Handler {
	counterRepository := repositories.NewCounterRepository(storage)
	gaugeRepository := repositories.NewGaugeRepository(storage)

	counterLogic := logic.NewCounterLogic(counterRepository)
	gaugeLogic := logic.NewGaugeLogic(gaugeRepository)

	updateMetricValueHTTPHandler := handlers.NewUpdateMetricValueHTTPHandler(
		counterLogic,
		gaugeLogic,
	)

	getMetricValueHTTPHandler := handlers.NewGetMetricValueHTTPHandler(
		counterRepository,
		gaugeRepository,
	)

	getAllMetricsHTTPHandler := handlers.NewGetAllMetricsHTTPHandler(
		storage,
	)

	router := chi.NewRouter()

	router.Post(protocol.UpdateMetricValueURL, updateMetricValueHTTPHandler.ServeHTTP)
	router.Get(protocol.GetMetricValueURL, getMetricValueHTTPHandler.ServeHTTP)
	router.Get(protocol.GetAllMetricsURL, getAllMetricsHTTPHandler.ServeHTTP)

	return router
}
