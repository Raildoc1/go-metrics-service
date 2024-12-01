package agent

import (
	"go-metrics-service/internal/agent/config"
	"go-metrics-service/internal/agent/metrics/collectors"
	"go-metrics-service/internal/agent/metrics/senders"
)

func Run(cfg config.Config) {
	mc := collectors.NewRuntimeMetricsCollector()
	ms := senders.NewMetricsSender(mc, cfg.ServerAddress)
	mc.StartPolling(cfg.PollingFreq)
	ms.StartSendingMetrics(cfg.PollingFreq, cfg.SendingFreq)
	select {}
}
