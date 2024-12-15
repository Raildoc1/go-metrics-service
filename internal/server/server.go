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

	updateMetricValueHTTPHandler := middleware.WithLogger(handlers.NewUpdateMetricValueHandler(counterLogic, gaugeLogic, logger), logger)
	updateJsonHTTPHandler := middleware.WithLogger(handlers.NewUpdateMetricValueJsonHandler(counterLogic, gaugeLogic, logger), logger)
	getMetricValueHTTPHandler := middleware.WithLogger(handlers.NewGetMetricValueTextHandler(rep, rep, logger), logger)
	getMetricValueJsonHandler := middleware.WithLogger(handlers.NewGetMetricValueJsonHandler(rep, rep, logger), logger)
	getAllMetricsHTTPHandler := middleware.WithLogger(handlers.NewGetAllMetrics(storage, logger), logger)

	router := chi.NewRouter()

	router.Post(protocol.UpdateJsonURL, updateJsonHTTPHandler.ServeHTTP)
	router.Post(protocol.UpdateMetricValueURL, updateMetricValueHTTPHandler.ServeHTTP)
	router.Post(protocol.GetMetricValueJsonURL, getMetricValueJsonHandler.ServeHTTP)
	router.Get(protocol.GetMetricValueURL, getMetricValueHTTPHandler.ServeHTTP)
	router.Get(protocol.GetAllMetricsURL, getAllMetricsHTTPHandler.ServeHTTP)

	return router
}
