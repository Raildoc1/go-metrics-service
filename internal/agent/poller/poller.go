package poller

import (
	"fmt"
	storagePkg "go-metrics-service/internal/agent/storage"
	gohelpers2 "go-metrics-service/pkg/gohelpers"
	"math/rand"
	"runtime"
	"time"

	"go.uber.org/zap"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

const (
	pollCountMetricName = "PollCount"
)

type Poller struct {
	storage *storagePkg.Storage
	logger  *zap.Logger
	doneCh  chan struct{}
}

func New(storage *storagePkg.Storage, logger *zap.Logger) *Poller {
	return &Poller{
		storage: storage,
		logger:  logger,
		doneCh:  make(chan struct{}),
	}
}

func (p *Poller) Start(interval time.Duration) chan error {
	errCh1 := gohelpers2.StartTickerProcess(p.doneCh, p.CollectRuntimeMetrics, interval)
	errCh2 := gohelpers2.StartTickerProcess(p.doneCh, p.CollectGopsutilMemMetrics, interval)
	errCh3 := gohelpers2.StartTickerProcess(p.doneCh, p.CollectGopsutilCPUMetrics, interval)

	return gohelpers2.AggregateErrors(errCh1, errCh2, errCh3)
}

func (p *Poller) Stop() {
	close(p.doneCh)
}

func (p *Poller) PollProcess(f func(), interval time.Duration, errCh chan error) {
	defer close(errCh)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			f()
		case <-p.doneCh:
			return
		}
	}
}

func (p *Poller) CollectRuntimeMetrics() error {
	runtimeMetrics := runtime.MemStats{}
	runtime.ReadMemStats(&runtimeMetrics)

	newPollCount := int64(1)
	if val, ok := p.storage.GetCounter(pollCountMetricName); ok {
		newPollCount = val + 1
	}

	p.storage.SetCounter(pollCountMetricName, newPollCount)

	p.storage.SetGauges(map[string]float64{
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
		"RandomValue":   rand.Float64(),
	})
	return nil
}

func (p *Poller) CollectGopsutilCPUMetrics() error {
	cpuInfos, err := cpu.Percent(0, true)
	if err != nil {
		return fmt.Errorf("failed to get cpu info: %w", err)
	}
	cpuValues := make(map[string]float64)
	for i, cpuInfo := range cpuInfos {
		cpuValues[fmt.Sprintf("CpuUtilization%v", i)] = cpuInfo
	}
	p.storage.SetGauges(cpuValues)
	return nil
}

func (p *Poller) CollectGopsutilMemMetrics() error {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to get mem info: %w", err)
	}
	p.storage.SetGauges(map[string]float64{
		"FreeMemory":  float64(memInfo.Free),
		"TotalMemory": float64(memInfo.Total),
	})
	return nil
}
