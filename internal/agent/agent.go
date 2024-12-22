package agent

import (
	"go-metrics-service/internal/agent/config"
	metricsCollector "go-metrics-service/internal/agent/metrics/collector"
	metricsSender "go-metrics-service/internal/agent/metrics/sender"
	metricsRequester "go-metrics-service/internal/agent/requester"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func Run(cfg config.Config, logger *zap.Logger) {
	collector := metricsCollector.New()
	requester := metricsRequester.New(cfg.ServerAddress, logger)
	sender := metricsSender.New(collector, requester)

	lifecycle(cfg, collector, sender)
}

func lifecycle(cfg config.Config, collector *metricsCollector.MetricsCollector, sender *metricsSender.MetricsSender) {
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(
		cancelChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGABRT,
	)

	pollTicker := time.NewTicker(cfg.PollingInterval)
	defer pollTicker.Stop()

	sendingTicker := time.NewTicker(cfg.SendingInterval)
	defer sendingTicker.Stop()

	for {
		select {
		case <-cancelChan:
			return
		case <-pollTicker.C:
			collector.Poll()
		case <-sendingTicker.C:
			send(collector, sender)
		}
	}
}

func send(collector *metricsCollector.MetricsCollector, sender *metricsSender.MetricsSender) {
	sender.TrySendRuntimeMetrics()
	sender.TrySendRandomValue()
	if ok := sender.TrySendPollCount(); ok {
		collector.FlushPollsCount()
	}
}
