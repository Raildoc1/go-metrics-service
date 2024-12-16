package main

import (
	"go-metrics-service/cmd/server/config"
	"go-metrics-service/internal/common/logging"
	"go-metrics-service/internal/server"
	"go-metrics-service/internal/server/data/storage/memory"
	"log"
	"net/http"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	memStorage := memory.NewMemStorage()
	logger, err := logging.CreateZapLogger(!cfg.Production)
	if err != nil {
		log.Fatal(err)
	}
	defer func(logger *zap.SugaredLogger) {
		err := logger.Sync()
		if err != nil {
			log.Fatal(err)
		}
	}(logger)
	handler := server.NewServer(memStorage, logger)
	err = http.ListenAndServe(cfg.ServerAddress, handler)
	if err != nil {
		logger.Error(err)
		return
	}
}
