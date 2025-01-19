package main

import (
	"encoding/json"
	"go-metrics-service/cmd/server/config"
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
	"log"
	"os"
	"os/signal"
	"syscall"

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
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			log.Println(err)
		}
	}(logger)

	jsCfg, err := json.MarshalIndent(cfg, "", "    ") //nolint:musttag // marshalling for debug
	if err != nil {
		logger.Error("Failed to marshal configuration", zap.Error(err))
		return
	}
	logger.Sugar().Infoln("Configuration: ", string(jsCfg))

	var pingables []handlers.Pingable
	var rep server.Repository
	var tm server.TransactionManager

	switch {
	case cfg.Database.ConnectionString != "":
		dbFactory := database.NewPgxDatabaseFactory(cfg.Database)
		dbStorage, err := dbstorage.New(dbFactory, cfg.Database.RetryAttempts, logger)
		if err != nil {
			logger.Error("Failed to create database storage", zap.Error(err))
			return
		}
		defer dbStorage.Close()
		rep = dbrepository.New(dbStorage, logger)
		tm = dbstorage.NewTransactionsManager(dbStorage, logger)
	case cfg.BackupMemStorage.Backup.FilePath != "":
		backupMemStorage, err := backupmemstorage.New(cfg.BackupMemStorage, logger)
		if err != nil {
			logger.Error("Failed to create memory storage", zap.Error(err))
			return
		}
		defer backupMemStorage.Stop()
		rep = memrepository.New(backupMemStorage, logger)
		tm = storages.NewDummyTransactionsManager()
	default:
		memStorage := memstorage.New(logger)
		rep = memrepository.New(memStorage, logger)
		tm = storages.NewDummyTransactionsManager()
	}

	srv := server.New(cfg.Server, rep, tm, pingables, logger)
	defer srv.Close()

	lifecycle(logger)

	logger.Info("Shutting down...")
}

func lifecycle(logger *zap.Logger) {
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
		logger.Info("Shutting down...")
		return
	}
}
