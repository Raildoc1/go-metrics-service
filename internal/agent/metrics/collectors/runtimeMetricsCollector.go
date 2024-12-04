package collectors

import (
	"errors"
	"runtime"
	"sync"
	"time"
)

type RuntimeMetricsCollector struct {
	runtimeMetrics runtime.MemStats
	startedMutex   sync.Mutex
	started        bool
}

func NewRuntimeMetricsCollector() *RuntimeMetricsCollector {
	return &RuntimeMetricsCollector{
		runtimeMetrics: runtime.MemStats{},
	}
}

func (mc *RuntimeMetricsCollector) StartPolling(interval time.Duration) error {
	mc.startedMutex.Lock()
	defer mc.startedMutex.Unlock()
	if mc.started {
		return errors.New("runtime metrics collector already started")
	}
	go func() {
		for {
			runtime.ReadMemStats(&mc.runtimeMetrics)
			time.Sleep(interval)
		}
	}()
	mc.started = true
	return nil
}

func (mc *RuntimeMetricsCollector) GetMetrics() runtime.MemStats {
	return mc.runtimeMetrics
}
