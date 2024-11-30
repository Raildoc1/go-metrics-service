package main

import (
	"go-metrics-service/cmd/server/config"
	"go-metrics-service/internal/server"
	"go-metrics-service/internal/server/data/storage"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	memStorage := storage.NewMemStorage()
	handler := server.NewServer(memStorage)
	err = http.ListenAndServe(cfg.ServerAddress, handler)
	if err != nil {
		log.Fatal(err)
	}
}
