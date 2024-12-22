package sender

import (
	"math/rand"
	"runtime"

	"go.uber.org/zap"
)

type requester interface {
	SendCounterDelta(metricName string, delta int64) error
	SendGauge(metricName string, value float64) error
}

type metricsStorage interface {
	GetRuntimeMetrics() runtime.MemStats
	GetPollsCount() int64
	FlushPollsCount()
}

type MetricsSender struct {
	storage   metricsStorage
	requester requester
	logger    *zap.Logger
}

func New(
	metricsStorage metricsStorage,
	requester requester,
	logger *zap.Logger,
) *MetricsSender {
	return &MetricsSender{
		storage:   metricsStorage,
		requester: requester,
		logger:    logger,
	}
}

func (ms *MetricsSender) TrySendRuntimeMetrics() {
	runtimeMetrics := ms.storage.GetRuntimeMetrics()
	runtimeMetricsMap := map[string]float64{
		"Alloc":         float64(runtimeMetrics.Alloc),
		"BuckHashSys":   float64(runtimeMetrics.BuckHashSys),
		"Frees":         float64(runtimeMetrics.Frees),
		"GCCPUFraction": runtimeMetrics.GCCPUFraction,
		"GCSys":         float64(runtimeMetrics.GCSys),
		"HeapAlloc":     float64(runtimeMetrics.HeapAlloc),
		"HeapIdle":      float64(runtimeMetrics.HeapIdle),
		"HeapInuse":     float64(runtimeMetrics.HeapInuse),
		"HeapObjects":   float64(runtimeMetrics.HeapObjects),
		"HeapReleased":  float64(runtimeMetrics.HeapReleased),
		"HeapSys":       float64(runtimeMetrics.HeapSys),
		"LastGC":        float64(runtimeMetrics.LastGC),
		"Lookups":       float64(runtimeMetrics.Lookups),
		"MCacheInuse":   float64(runtimeMetrics.MCacheInuse),
		"MCacheSys":     float64(runtimeMetrics.MCacheSys),
		"MSpanInuse":    float64(runtimeMetrics.MSpanInuse),
		"MSpanSys":      float64(runtimeMetrics.MSpanSys),
		"Mallocs":       float64(runtimeMetrics.Mallocs),
		"NextGC":        float64(runtimeMetrics.NextGC),
		"NumForcedGC":   float64(runtimeMetrics.NumForcedGC),
		"NumGC":         float64(runtimeMetrics.NumGC),
		"OtherSys":      float64(runtimeMetrics.OtherSys),
		"PauseTotalNs":  float64(runtimeMetrics.PauseTotalNs),
		"StackInuse":    float64(runtimeMetrics.StackInuse),
		"StackSys":      float64(runtimeMetrics.StackSys),
		"Sys":           float64(runtimeMetrics.Sys),
		"TotalAlloc":    float64(runtimeMetrics.TotalAlloc),
	}

	for key, value := range runtimeMetricsMap {
		ms.trySendGauge(key, value)
	}
}

func (ms *MetricsSender) TrySendRandomValue() {
	ms.trySendGauge("RandomValue", rand.Float64())
}

func (ms *MetricsSender) TrySendPollCount() bool {
	return ms.trySendCounterDelta("PollCount", ms.storage.GetPollsCount())
}

func (ms *MetricsSender) trySendGauge(key string, value float64) {
	err := ms.requester.SendGauge(key, value)
	if err != nil {
		ms.logger.Error("Failed to send gauge",
			zap.String("key", key),
			zap.Float64("value", value),
			zap.Error(err),
		)
	}
}

func (ms *MetricsSender) trySendCounterDelta(key string, value int64) bool {
	err := ms.requester.SendCounterDelta(key, value)
	if err != nil {
		ms.logger.Error("Failed to send counter",
			zap.String("key", key),
			zap.Int64("delta", value),
			zap.Error(err),
		)
	}
	return err == nil
}
