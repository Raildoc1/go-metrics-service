package main

import (
	"go-metrics-service/cmd/agent/config"
	"go-metrics-service/internal/agent"
	"go-metrics-service/internal/common/logging"
	"log"
	"os"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logFile, err := os.Create("./agent.log")
	if err != nil {
		log.Fatal(err)
	}
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			log.Println(err)
		}
	}(logFile)
	logger := logging.CreateZapLogger(!cfg.Production, logFile)
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			log.Println(err)
		}
	}(logger)

	agent.Run(cfg.Agent, logger)
}
