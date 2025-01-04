package main

import (
	"go-metrics-service/cmd/server/config"
	"go-metrics-service/internal/common/logging"
	"go-metrics-service/internal/server"
	"go-metrics-service/internal/server/data/backupmemstorage"
	"go-metrics-service/internal/server/data/dbstorage"
	"go-metrics-service/internal/server/data/memstorage"
	"go-metrics-service/internal/server/database"
	"go-metrics-service/internal/server/handlers"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := logging.CreateZapLogger(!cfg.Production)
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			log.Println(err)
		}
	}(logger)

	var pingables []handlers.Pingable
	var storage server.Storage

	switch {
	case cfg.Database.ConnectionString != "":
		dbFactory := database.NewPgxDatabaseFactory(cfg.Database)
		dbStorage, err := dbstorage.New(dbFactory, logger)
		if err != nil {
			logger.Error("Failed to create database storage", zap.Error(err))
			return
		}
		defer dbStorage.Close()
		storage = dbStorage
	case cfg.BackupMemStorage.Backup.FilePath != "":
		backupMemStorage, err := backupmemstorage.New(cfg.BackupMemStorage, logger)
		if err != nil {
			logger.Error("Failed to create memory storage", zap.Error(err))
			return
		}
		defer backupMemStorage.Stop()
		storage = backupMemStorage
	default:
		storage = memstorage.NewMemStorage(logger)
	}

	srv := server.New(cfg.Server, storage, pingables, logger)
	defer srv.Close()

	lifecycle()
}

func lifecycle() {
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(
		cancelChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGABRT,
	)

	for range cancelChan {
	}
}
