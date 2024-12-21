package main

import (
	"go-metrics-service/cmd/agent/config"
	"go-metrics-service/internal/agent"
	"go-metrics-service/internal/common/logging"
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

	agent.Run(cfg.Agent, logger)
}
