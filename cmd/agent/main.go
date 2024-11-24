package main

import (
	"go-metrics-service/cmd/agent/metricsSender"
	"go-metrics-service/cmd/agent/runtimeMetricsCollector"
	"time"
)

func main() {
	mc := runtimeMetricsCollector.NewRuntimeMetricsCollector()
	ms := metricsSender.NewMetricsSender(mc, "localhost:8080")
	mc.StartPolling(2 * time.Second)
	ms.StartSendingMetrics(10 * time.Second)
	select {}
}
