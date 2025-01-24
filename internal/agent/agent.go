package agent

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"go-metrics-service/internal/agent/config"
	pollerPkg "go-metrics-service/internal/agent/poller"
	senderPkg "go-metrics-service/internal/agent/sender"
	storagePkg "go-metrics-service/internal/agent/storage"
	"go-metrics-service/internal/common/timeutils"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func Run(cfg config.Config, logger *zap.Logger) {
	storage := storagePkg.New()
	poller := pollerPkg.New(storage)
	hash := hmac.New(sha256.New, []byte(cfg.SHA256Key))
	sender := senderPkg.New(cfg.ServerAddress, storage, logger, hash)

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
			_ = timeutils.Retry(
				context.Background(),
				cfg.RetryAttempts,
				func(ctx context.Context) error {
					return sender.Send()
				},
				func(err error) (needRetry bool) {
					needRetry = errors.Is(err, senderPkg.ErrServerUnavailable)
					if needRetry {
						logger.Error("sending failed", zap.Error(err))
					}
					return needRetry
				},
			)
		}
	}
}
