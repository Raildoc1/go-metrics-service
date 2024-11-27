package main

import (
	"github.com/go-chi/chi/v5"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data/repositories"
	"go-metrics-service/internal/server/data/storage"
	"go-metrics-service/internal/server/handlers"
	"go-metrics-service/internal/server/logic"
	"net/http"
)

func main() {
	memStorage := storage.NewMemStorage()

	counterRepository := repositories.NewCounterRepository(memStorage)
	gaugeRepository := repositories.NewGaugeRepository(memStorage)

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

	r := chi.NewRouter()

	r.Post(protocol.UpdateMetricValueUrl, updateMetricValueHTTPHandler.ServeHTTP)
	r.Get(protocol.GetMetricValueUrl, getMetricValueHTTPHandler.ServeHTTP)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
