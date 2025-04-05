package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-metrics-service/cmd/server/config"
	"go-metrics-service/internal/common/hashing"
	"go-metrics-service/internal/common/logging"
	"go-metrics-service/internal/server"
	"go-metrics-service/internal/server/data/repositories/dbrepository"
	"go-metrics-service/internal/server/data/repositories/memrepository"
	"go-metrics-service/internal/server/data/storages"
	"go-metrics-service/internal/server/data/storages/backupmemstorage"
	"go-metrics-service/internal/server/data/storages/dbstorage"
	"go-metrics-service/internal/server/data/storages/memstorage"
	"go-metrics-service/internal/server/database"
	"go-metrics-service/internal/server/handlers"
	"go-metrics-service/internal/server/middleware"
	"log"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"go.uber.org/zap"
)

func main() {
	log.Println("Starting server...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := logging.CreateZapLogger(!cfg.Production).
		With(zap.String("source", "server"))
	defer syncZapLogger(logger)

	jsCfg, err := json.MarshalIndent(cfg, "", "    ") //nolint:musttag // marshalling for debug
	if err != nil {
		logger.Error("Failed to marshal configuration", zap.Error(err))
		return
	}
	logger.Sugar().Infoln("Configuration: ", string(jsCfg))

	if err := run(&cfg, logger); err != nil {
		logger.Error("Server shutdown with error", zap.Error(err))
	} else {
		logger.Info("Server shutdown gracefully")
	}
}

func run(cfg *config.Config, logger *zap.Logger) error {
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

	context.AfterFunc(ctx, func() {
		timeoutCtx, cancelCtx := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancelCtx()

		<-timeoutCtx.Done()
		log.Fatal("failed to gracefully shutdown the server")
	})

	var pingables []handlers.Pingable
	var rep server.Repository
	var tm server.TransactionManager

	switch {
	case cfg.Database.ConnectionString != "":
		dbFactory := database.NewPgxDatabaseFactory(cfg.Database)
		dbStorage, err := dbstorage.New(dbFactory, cfg.Database.RetryAttempts, logger)
		if err != nil {
			return fmt.Errorf("failed to create database storage: %w", err)
		}
		g.Go(func() error {
			defer logger.Info("Closing DB Storage")
			<-ctx.Done()
			dbStorage.Close()
			return nil
		})
		rep = dbrepository.New(dbStorage, logger)
		tm = dbstorage.NewTransactionsManager(dbStorage, logger)
	case cfg.BackupMemStorage.Backup.FilePath != "":
		backupMemStorage, err := backupmemstorage.New(cfg.BackupMemStorage, logger)
		if err != nil {
			return fmt.Errorf("failed to create memory storage: %w", err)
		}
		g.Go(func() error {
			defer logger.Info("Stopping mem-storage backup")
			<-ctx.Done()
			defer backupMemStorage.Stop()
			return nil
		})
		rep = memrepository.New(backupMemStorage, logger)
		tm = storages.NewDummyTransactionsManager()
	default:
		memStorage := memstorage.New(logger)
		rep = memrepository.New(memStorage, logger)
		tm = storages.NewDummyTransactionsManager()
	}

	var hashFactory middleware.HashFactory = nil
	if cfg.SHA256Key != "" {
		hashFactory = hashing.NewHMAC(cfg.SHA256Key)
	}

	srv := server.New(cfg.Server, rep, tm, hashFactory, pingables, logger)

	g.Go(func() error {
		if err := srv.Run(); err != nil {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		defer logger.Info("Shutting down server")
		<-ctx.Done()
		if err := srv.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("goroutine error occured: %w", err)
	}

	return nil
}

func syncZapLogger(logger *zap.Logger) {
	err := logger.Sync()
	if err != nil {
		log.Println(err)
	}
}
