package collector

import (
	"runtime"
)

type MetricsCollector struct {
	runtimeMetrics runtime.MemStats
	pollsCount     int64
}

func New() *MetricsCollector {
	return &MetricsCollector{
		runtimeMetrics: runtime.MemStats{},
		pollsCount:     0,
	}
}

func (mc *MetricsCollector) Poll() {
	runtime.ReadMemStats(&mc.runtimeMetrics)
	mc.pollsCount++
}

func (mc *MetricsCollector) GetRuntimeMetrics() runtime.MemStats {
	return mc.runtimeMetrics
}

func (mc *MetricsCollector) GetPollsCount() int64 {
	return mc.pollsCount
}

func (mc *MetricsCollector) FlushPollsCount() {
	mc.pollsCount = 0
}
