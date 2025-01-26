package agent

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"go-metrics-service/internal/agent/config"
	pollerPkg "go-metrics-service/internal/agent/poller"
	senderPkg "go-metrics-service/internal/agent/sender"
	storagePkg "go-metrics-service/internal/agent/storage"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"go.uber.org/zap"
)

func Run(cfg *config.Config, logger *zap.Logger) error {
	storage := storagePkg.New()
	poller := pollerPkg.New(storage, logger)
	hash := hmac.New(sha256.New, []byte(cfg.SHA256Key))
	sender := senderPkg.New(cfg.ServerAddress, storage, logger, hash)

	rootCtx, cancelCtx := signal.NotifyContext(
		context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGABRT,
	)
	defer cancelCtx()

	g, ctx := errgroup.WithContext(rootCtx)

	g.Go(func() error {
		errCh := poller.Start(cfg.PollingInterval)
		for err := range errCh {
			logger.Error("poller error", zap.Error(err))
		}
		return nil
	})

	g.Go(func() error {
		defer logger.Info("Poller stopped")
		<-ctx.Done()
		poller.Stop()
		return nil
	})

	g.Go(func() error {
		errCh := sender.Start(cfg.SendingInterval, cfg.RateLimit)
		for err := range errCh {
			logger.Error("sender error", zap.Error(err))
		}
		return nil
	})

	g.Go(func() error {
		defer logger.Info("Sender stopped")
		<-ctx.Done()
		sender.Stop()
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("goroutine error occured: %w", err)
	}

	return nil
}
