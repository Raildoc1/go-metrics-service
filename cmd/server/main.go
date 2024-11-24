package main

import (
	"go-metrics-service/cmd/server/data/repositories"
	"go-metrics-service/cmd/server/data/storage"
	"go-metrics-service/cmd/server/handlers"
	"go-metrics-service/cmd/server/logic"
	"net/http"
)

func main() {
	memStorage := storage.NewMemStorage()

	counterRepository := repositories.NewCounterRepository(memStorage)
	gaugeRepository := repositories.NewGaugeRepository(memStorage)

	counterLogic := logic.NewCounterLogic(counterRepository)
	gaugeLogic := logic.NewGaugeLogic(gaugeRepository)

	mux := http.NewServeMux()

	mux.Handle("/update/gauge/{key}/{value}", handlers.NewGaugeHTTPHandler(gaugeLogic))
	mux.Handle("/update/counter/{key}/{value}", handlers.NewCounterHTTPHandler(counterLogic))
	mux.Handle("/update/{typename}/{key}/{value}", &handlers.DummyHTTPHandler{})

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
