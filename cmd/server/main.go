package main

import (
	"go-metrics-service/cmd/server/config"
	"go-metrics-service/internal/server"
	"go-metrics-service/internal/server/data/storage/memory"
	"go-metrics-service/internal/server/logger"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	memStorage := memory.NewMemStorage()
	err = logger.CreateLogger(cfg.Development)
	if err != nil {
		log.Fatal(err)
	}
	handler := server.NewServer(memStorage)
	err = http.ListenAndServe(cfg.ServerAddress, handler)
	if err != nil {
		log.Fatal(err)
	}
	err = logger.Sync()
	if err != nil {
		log.Fatal(err)
	}
}
