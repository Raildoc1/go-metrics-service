package main

import (
	"go-metrics-service/cmd/server/config"
	"go-metrics-service/internal/common/logging"
	"go-metrics-service/internal/server"
	"log"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger, err := logging.CreateZapLogger(!cfg.Production)
	if err != nil {
		log.Fatal(err)
	}
	defer func(logger *zap.SugaredLogger) {
		err := logger.Sync()
		if err != nil {
			log.Println(err)
		}
	}(logger)

	server.Run(cfg.Server, logger)
}
