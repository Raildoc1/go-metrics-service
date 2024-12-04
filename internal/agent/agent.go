package agent

import (
	"go-metrics-service/internal/agent/config"
	"go-metrics-service/internal/agent/metrics/collectors"
	"go-metrics-service/internal/agent/metrics/senders"
)

func Run(cfg config.Config) error {
	mc := collectors.NewRuntimeMetricsCollector()
	ms := senders.NewMetricsSender(mc, cfg.ServerAddress)
	err := mc.StartPolling(cfg.PollingFreq)
	if err != nil {
		return err
	}
	err = ms.StartSendingMetrics(cfg.PollingFreq, cfg.SendingFreq)
	if err != nil {
		return err
	}
	select {}
}
