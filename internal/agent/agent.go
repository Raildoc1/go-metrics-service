package agent

import (
	"go-metrics-service/internal/agent/config"
	pollerPkg "go-metrics-service/internal/agent/poller"
	requesterPkg "go-metrics-service/internal/agent/requester"
	senderPkg "go-metrics-service/internal/agent/sender"
	storagePkg "go-metrics-service/internal/agent/storage"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func Run(cfg config.Config, logger *zap.Logger) {
	storage := storagePkg.New()
	poller := pollerPkg.New(storage)
	requester := requesterPkg.New(cfg.ServerAddress, logger)
	sender := senderPkg.New(storage, requester)

	lifecycle(cfg, poller, sender, logger)
}

func lifecycle(cfg config.Config, poller *pollerPkg.Poller, sender *senderPkg.Sender, logger *zap.Logger) {
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
			poller.Poll()
		case <-sendingTicker.C:
			if err := sender.Send(); err != nil {
				logger.Error("sending failed", zap.Error(err))
			}
		}
	}
}
