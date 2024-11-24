package metricsSender

import (
	"fmt"
	"go-metrics-service/cmd/agent/runtimeMetricsCollector"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type MetricsSender struct {
	runtimeMetricsCollector *runtimeMetricsCollector.RuntimeMetricsCollector
	started                 bool
	host                    string
}

func NewMetricsSender(
	runtimeMetricsCollector *runtimeMetricsCollector.RuntimeMetricsCollector,
	host string,
) *MetricsSender {
	return &MetricsSender{
		runtimeMetricsCollector: runtimeMetricsCollector,
		host:                    host,
	}
}

func (ms *MetricsSender) StartSendingMetrics(interval time.Duration) {
	if ms.started {
		panic("already started")
	}
	go func() {
		for {
			ms.sendMetrics()
			time.Sleep(interval)
		}
	}()
	ms.started = true
}

func (ms *MetricsSender) sendMetrics() {
	runtimeMetrics := ms.runtimeMetricsCollector.GetMetrics()
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
		ms.sendGaugeWithErrorHandling(key, value)
	}

	ms.sendCounterDeltaWithErrorHandling("PollCount", 1)
	ms.sendGaugeWithErrorHandling("RandomValue", rand.Float64())
}

func (ms *MetricsSender) sendCounterDeltaWithErrorHandling(metricName string, delta int64) {
	resp, err := ms.sendCounterDelta(metricName, delta)
	if err != nil {
		fmt.Printf("%s: %s\n", metricName, err.Error())
	} else if resp.StatusCode != http.StatusOK {
		fmt.Printf("%s: status %s\n", metricName, resp.StatusCode)
	}
}

func (ms *MetricsSender) sendGaugeWithErrorHandling(metricName string, value float64) {
	resp, err := ms.sendGauge(metricName, value)
	if err != nil {
		fmt.Printf("%s: %s\n", metricName, err.Error())
	} else if resp.StatusCode != http.StatusOK {
		fmt.Printf("%s: status %s\n", metricName, resp.StatusCode)
	}
}

func (ms *MetricsSender) sendCounterDelta(metricName string, delta int64) (*http.Response, error) {
	return http.Post(
		buildUpdateRequest(
			ms.host,
			"counter",
			metricName,
			strconv.FormatInt(delta, 10),
		),
		"text/plain",
		nil,
	)
}

func (ms *MetricsSender) sendGauge(metricName string, value float64) (*http.Response, error) {
	return http.Post(
		buildUpdateRequest(
			ms.host,
			"gauge",
			metricName,
			strconv.FormatFloat(value, 'f', -1, 64),
		),
		"text/plain",
		nil,
	)
}

func buildUpdateRequest(host, metricType, metricKey, metricValue string) string {
	return fmt.Sprintf("http://%s/update/%s/%s/%s", host, metricType, metricKey, metricValue)
}
