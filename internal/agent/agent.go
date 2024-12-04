package agent

import (
	"go-metrics-service/internal/agent/config"
	"go-metrics-service/internal/agent/metrics/collecting"
	"go-metrics-service/internal/agent/metrics/sending"
	"go-metrics-service/internal/agent/requesting"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(cfg config.Config) error {
	collector := collecting.NewMetricsCollector()
	requester := requesting.NewRequester(cfg.ServerAddress)
	sender := sending.NewMetricsSender(collector, requester)

	lifecycle(cfg, collector, sender)

	return nil
}

func lifecycle(cfg config.Config, collector *collecting.MetricsCollector, sender *sending.MetricsSender) {
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
