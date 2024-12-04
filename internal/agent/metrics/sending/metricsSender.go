package sending

import (
	"log"
	"math/rand"
	"runtime"
)

type Requester interface {
	SendCounterDelta(metricName string, delta int64) error
	SendGauge(metricName string, value float64) error
}

type MetricsStorage interface {
	GetRuntimeMetrics() runtime.MemStats
	GetPollsCount() int64
	FlushPollsCount()
}

type MetricsSender struct {
	storage   MetricsStorage
	requester Requester
}

func NewMetricsSender(
	metricsStorage MetricsStorage,
	requester Requester,
) *MetricsSender {
	return &MetricsSender{
		storage:   metricsStorage,
		requester: requester,
	}
}

func (ms *MetricsSender) Send() {
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
		_ = ms.sendGauge(key, value)
	}

	_ = ms.sendGauge("RandomValue", rand.Float64())

	if err := ms.sendCounterDelta("PollCount", ms.storage.GetPollsCount()); err != nil {
		ms.storage.FlushPollsCount()
	}
}

func (ms *MetricsSender) sendGauge(key string, value float64) error {
	err := ms.requester.SendGauge(key, value)
	if err != nil {
		log.Printf("Error sending gauge '%s': %v", key, err)
	}
	return err
}

func (ms *MetricsSender) sendCounterDelta(key string, value int64) error {
	err := ms.requester.SendCounterDelta(key, value)
	if err != nil {
		log.Printf("Error sending counter '%s': %v", key, err)
	}
	return err
}
