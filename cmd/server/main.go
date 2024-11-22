package main

import (
	"go-metrics-service/cmd/server/handlers"
	"go-metrics-service/cmd/server/logic"
	"go-metrics-service/cmd/server/storage"
	"net/http"
)

func main() {
	memStorage := storage.NewMemStorage()

	counterLogic := logic.NewCounterLogic(memStorage)
	gaugeLogic := logic.NewGaugeLogic(memStorage)

	mux := http.NewServeMux()

	mux.Handle("/update/gauge/{key}/{value}", handlers.NewGaugeHTTPHandler(gaugeLogic))
	mux.Handle("/update/counter/{key}/{value}", handlers.NewCounterHTTPHandler(counterLogic))
	mux.Handle("/update/{typename}/{key}/{value}", &handlers.DummyHTTPHandler{})

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
