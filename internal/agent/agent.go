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
)

func Run(cfg config.Config) error {
	collector := metricsCollector.New()
	requester := metricsRequester.New(cfg.ServerAddress)
	sender := metricsSender.New(collector, requester)

	lifecycle(cfg, collector, sender)

	return nil
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
		syscall.SIGKILL,
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
			sender.Send()
		}
	}
}
