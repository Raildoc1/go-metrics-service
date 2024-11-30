package agent

import (
	"go-metrics-service/internal/agent/config"
	"go-metrics-service/internal/agent/metrics/collectors"
	"go-metrics-service/internal/agent/metrics/senders"
)

func Run(config config.Config) {
	mc := collectors.NewRuntimeMetricsCollector()
	ms := senders.NewMetricsSender(mc, config.ServerAddress)
	mc.StartPolling(config.PollingFreq)
	ms.StartSendingMetrics(config.PollingFreq, config.SendingFreq)
	select {}
}
