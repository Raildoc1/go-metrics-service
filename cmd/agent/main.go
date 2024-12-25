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
	logger := logging.CreateZapLogger(!cfg.Production)
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			log.Println(err)
		}
	}(logger)

	agent.Run(cfg.Agent, logger)
}
