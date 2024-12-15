package server

import (
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data/repository"
	"go-metrics-service/internal/server/handlers"
	"go-metrics-service/internal/server/logic/counter"
	"go-metrics-service/internal/server/logic/gauge"
	"go-metrics-service/internal/server/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Logger interface {
	middleware.Logger
	handlers.Logger
}

func NewServer(storage repository.Storage, logger Logger) http.Handler {
	rep := repository.New(storage)

	counterLogic := counter.New(rep)
	gaugeLogic := gauge.New(rep)

	updateMetricPathParamsHandler := middleware.WithLogger(
		handlers.NewUpdateMetricPathParams(counterLogic, gaugeLogic, logger),
		logger,
	)

	updateMetricHandler := middleware.WithLogger(
		handlers.NewUpdateMetric(counterLogic, gaugeLogic, logger),
		logger,
	)

	getMetricValuePathParamsHandler := middleware.WithLogger(
		handlers.NewGetMetricValuePathParams(rep, rep, logger),
		logger,
	)

	getMetricValueHandler := middleware.WithLogger(
		handlers.NewGetMetricValue(rep, rep, logger),
		logger,
	)

	getAllMetricsHandler := middleware.WithLogger(
		handlers.NewGetAllMetrics(storage, logger),
		logger,
	)

	router := chi.NewRouter()

	router.Post(protocol.UpdateMetricURL, updateMetricHandler.ServeHTTP)
	router.Post(protocol.UpdateMetricPathParamsURL, updateMetricPathParamsHandler.ServeHTTP)
	router.Post(protocol.GetMetricURL, getMetricValueHandler.ServeHTTP)
	router.Get(protocol.GetMetricPathParamsURL, getMetricValuePathParamsHandler.ServeHTTP)
	router.Get(protocol.GetAllMetricsURL, getAllMetricsHandler.ServeHTTP)

	return router
}
