package main

import (
	"go-metrics-service/cmd/agent/metrics/collectors"
	"go-metrics-service/cmd/agent/metrics/senders"
	"time"
)

func main() {
	mc := collectors.NewRuntimeMetricsCollector()
	ms := senders.NewMetricsSender(mc, "localhost:8080")
	mc.StartPolling(2 * time.Second)
	ms.StartSendingMetrics(2*time.Second, 10*time.Second)
	select {}
}
