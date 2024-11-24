package runtimeMetricsCollector

import (
	"runtime"
	"time"
)

type RuntimeMetricsCollector struct {
	runtimeMetrics runtime.MemStats
	started        bool
}

func NewRuntimeMetricsCollector() *RuntimeMetricsCollector {
	return &RuntimeMetricsCollector{
		runtimeMetrics: runtime.MemStats{},
	}
}

func (mc *RuntimeMetricsCollector) StartPolling(interval time.Duration) {
	if mc.started {
		panic("already started")
	}
	go func() {
		for {
			runtime.ReadMemStats(&mc.runtimeMetrics)
			time.Sleep(interval)
		}
	}()
	mc.started = true
}

func (mc *RuntimeMetricsCollector) GetMetrics() runtime.MemStats {
	return mc.runtimeMetrics
}
