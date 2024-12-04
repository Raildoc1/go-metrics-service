package main

import (
	"go-metrics-service/cmd/agent/config"
	"go-metrics-service/internal/agent"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	err = agent.Run(cfg.Agent)
	if err != nil {
		log.Fatal(err)
	}
}
