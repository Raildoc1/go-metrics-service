// Package agent contains composition root for agent
package agent

import (
	"context"
	"fmt"
	"go-metrics-service/internal/agent/config"
	pollerPkg "go-metrics-service/internal/agent/poller"
	senderPkg "go-metrics-service/internal/agent/sender"
	"go-metrics-service/internal/agent/sender/driver"
	storagePkg "go-metrics-service/internal/agent/storage"
	"go-metrics-service/internal/common/hashing"
	"go-metrics-service/pkg/ipdeterminer"
	"go-metrics-service/pkg/rsahelpers"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"go.uber.org/zap"
)

func Run(cfg *config.Config, logger *zap.Logger) error {
	ip, err := ipdeterminer.GetPreferredOutboundIP(logger)
	if err != nil {
		return fmt.Errorf("getting IP failed: %w", err)
	}

	storage := storagePkg.New()
	poller := pollerPkg.New(storage, logger)

	var hashFactory driver.HashFactory = nil
	if cfg.SHA256Key != "" {
		hashFactory = hashing.NewHMAC(cfg.SHA256Key)
	}

	var encoder driver.Encoder
	if cfg.RSAPublicKeyPem != nil {
		e, err := rsahelpers.NewOAEPEncoder(cfg.RSAPublicKeyPem)
		if err != nil {
			return err
		}
		encoder = e
	}

	var drv senderPkg.Driver = nil

	if cfg.GRPC != nil {
		d, err := driver.NewGrpcDriver(*cfg.GRPC)
		if err != nil {
			return err
		}
		drv = d
	} else {
		drv = driver.NewHTTPDriver(
			logger,
			storage,
			hashFactory,
			cfg.ServerAddress,
			encoder,
			ip,
		)
	}

	sender := senderPkg.New(cfg.RetryAttempts, storage, logger, drv)

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
		defer logger.Info("Poller errors handler stopped")
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
		defer logger.Info("Sender errors handler stopped")
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
