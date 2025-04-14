package main

import (
	"encoding/json"
	"fmt"
	"go-metrics-service/cmd/agent/config"
	"go-metrics-service/internal/agent"
	"go-metrics-service/internal/common/logging"
	"log"

	"go.uber.org/zap"
)

var buildVersion = "N/A"
var buildDate = "N/A"
var buildCommit = "N/A"

func main() {
	printBuildInfo()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := logging.CreateZapLogger(!cfg.Production).
		With(zap.String("source", "agent"))
	defer syncZapLogger(logger)

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

func syncZapLogger(logger *zap.Logger) {
	err := logger.Sync()
	if err != nil {
		log.Println(err)
	}
}

func printBuildInfo() {
	fmt.Printf("Build Version: %s\n", buildVersion)
	fmt.Printf("Build Date: %s\n", buildDate)
	fmt.Printf("Build Commit: %s\n", buildCommit)
}
