package agent

import (
	"go-metrics-service/internal/agent/metrics/collectors"
	"go-metrics-service/internal/agent/metrics/senders"
	"time"
)

func Run(serverAddress string, pollingFreq, sendingFreq time.Duration) {
	mc := collectors.NewRuntimeMetricsCollector()
	ms := senders.NewMetricsSender(mc, serverAddress)
	mc.StartPolling(pollingFreq)
	ms.StartSendingMetrics(pollingFreq, sendingFreq)
	select {}
}
