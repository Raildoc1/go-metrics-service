package main

import (
	"encoding/json"
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
	logger := logging.CreateZapLogger(!cfg.Production).
		With(zap.String("source", "agent"))
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			log.Println(err)
		}
	}(logger)

	jsCfg, err := json.MarshalIndent(cfg, "", "    ") //nolint:musttag // marshalling for debug
	if err != nil {
		logger.Error("Failed to marshal configuration", zap.Error(err))
		return
	}
	logger.Sugar().Infoln("Configuration: ", string(jsCfg))

	err = agent.Run(&cfg.Agent, logger)
	if err != nil {
		logger.Error("Agent error", zap.Error(err))
	}
}
