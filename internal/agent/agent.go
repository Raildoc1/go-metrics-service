// Package agent contains composition root for agent
package agent

import (
	"context"
	"fmt"
	"go-metrics-service/internal/agent/config"
	pollerPkg "go-metrics-service/internal/agent/poller"
	senderPkg "go-metrics-service/internal/agent/sender"
	storagePkg "go-metrics-service/internal/agent/storage"
	"go-metrics-service/internal/common/hashing"
	"go-metrics-service/pkg/rsahelpers"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"go.uber.org/zap"
)

func Run(cfg *config.Config, logger *zap.Logger) error {
	storage := storagePkg.New()
	poller := pollerPkg.New(storage, logger)

	var hashFactory senderPkg.HashFactory = nil
	if cfg.SHA256Key != "" {
		hashFactory = hashing.NewHMAC(cfg.SHA256Key)
	}

	var encoder senderPkg.Encoder
	if cfg.RSAPublicKeyPem != nil {
		e, err := rsahelpers.NewOAEPEncoder(cfg.RSAPublicKeyPem)
		if err != nil {
			return err
		}
		encoder = e
	}

	sender := senderPkg.New(
		cfg.ServerAddress,
		cfg.RetryAttempts,
		storage,
		logger,
		hashFactory,
		encoder,
	)

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
