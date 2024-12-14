package server

import (
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data/repository"
	"go-metrics-service/internal/server/handlers/getallmetrics"
	"go-metrics-service/internal/server/handlers/getmetricvalue"
	"go-metrics-service/internal/server/handlers/updatejson"
	"go-metrics-service/internal/server/handlers/updatemetricvalue"
	"go-metrics-service/internal/server/logic/counter"
	"go-metrics-service/internal/server/logic/gauge"
	"go-metrics-service/internal/server/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Logger interface {
	middleware.InfoLogger
	getallmetrics.Logger
	getmetricvalue.Logger
	updatemetricvalue.Logger
	updatejson.Logger
}

func NewServer(storage repository.Storage, logger Logger) http.Handler {
	rep := repository.New(storage)

	counterLogic := counter.New(rep)
	gaugeLogic := gauge.New(rep)

	updateMetricValueHTTPHandler := middleware.WithLogger(updatemetricvalue.New(counterLogic, gaugeLogic, logger), logger)
	updateJsonHTTPHandler := middleware.WithLogger(updatejson.New(counterLogic, gaugeLogic, logger), logger)
	getMetricValueHTTPHandler := middleware.WithLogger(getmetricvalue.New(rep, logger), logger)
	getAllMetricsHTTPHandler := middleware.WithLogger(getallmetrics.New(storage, logger), logger)

	router := chi.NewRouter()

	router.Post(protocol.UpdateJsonURL, updateJsonHTTPHandler.ServeHTTP)
	router.Post(protocol.UpdateMetricValueURL, updateMetricValueHTTPHandler.ServeHTTP)
	router.Get(protocol.GetMetricValueURL, getMetricValueHTTPHandler.ServeHTTP)
	router.Get(protocol.GetAllMetricsURL, getAllMetricsHTTPHandler.ServeHTTP)

	return router
}
