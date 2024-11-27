package main

import (
	"go-metrics-service/internal/server"
	"go-metrics-service/internal/server/data/storage"
	"net/http"
)

func main() {
	memStorage := storage.NewMemStorage()
	handler := server.NewServer(memStorage)
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		panic(err)
	}
}
