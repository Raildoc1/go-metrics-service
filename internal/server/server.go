package server

import (
	"context"
	"database/sql"
	"errors"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data/storage"
	"go-metrics-service/internal/server/database"
	"go-metrics-service/internal/server/handlers"
	"go-metrics-service/internal/server/logic"
	"go-metrics-service/internal/server/middleware"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Database interface {
	handlers.Database
}

func Run(cfg Config, logger *zap.Logger) {
	db, err := database.New(cfg.Database)
	if err != nil {
		logger.Error("failed to create database", zap.Error(err))
		return
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("failed to close database", zap.Error(err))
		}
	}(db)

	dbStorage, err := storage.NewDatabaseStorage(db, logger)
	if err != nil {
		logger.Error("failed to create database storage", zap.Error(err))
		return
	}

	srv := &http.Server{Addr: cfg.ServerAddress}
	srv.Handler = createMux(dbStorage, db, logger)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to start server", zap.Error(err))
		}
	}()

	lifecycle(cfg, logger, dbStorage)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("failed to gracefully shutdown", zap.Error(err))
	}
}

//func createMemStorage(cfg Config, logger *zap.Logger) (*storage.DatabaseStorage, error) {
//	if cfg.NeedRestore {
//		if _, err := os.Stat(cfg.FilePath); err == nil {
//			file, err := os.Open(cfg.FilePath)
//			if err != nil {
//				return nil, fmt.Errorf("failed to open file: %w", err)
//			}
//			ms, err := storage.LoadFrom(file, logger)
//			if err != nil {
//				return nil, fmt.Errorf("failed to load from file: %w", err)
//			}
//			logger.Info("data successfully restored", zap.String("path", cfg.FilePath))
//			return ms, nil
//		}
//	}
//	return storage.NewMemStorage(logger), nil
//}

func lifecycle(cfg Config, logger *zap.Logger, memStorage *storage.DatabaseStorage) {
	storeTicker := time.NewTicker(cfg.StoreInterval)

	cancelChan := make(chan os.Signal, 1)
	signal.Notify(
		cancelChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGABRT,
	)

	for {
		select {
		case <-storeTicker.C:
			// trySaveStorage(memStorage, cfg.FilePath, logger)
			break
		case <-cancelChan:
			// trySaveStorage(memStorage, cfg.FilePath, logger)
			return
		}
	}
}

//func trySaveStorage(memStorage *storage.DatabaseStorage, filePath string, logger *zap.Logger) {
//	if err := storage.SaveMemStorageToFile(memStorage, filePath, logger); err != nil {
//		logger.Error("failed to save to file", zap.Error(err))
//	} else {
//		logger.Info("successfully saved to file", zap.String("path", filePath))
//	}
//}

func createMux(memStorage *storage.DatabaseStorage, db Database, logger *zap.Logger) *chi.Mux {
	counterLogic := logic.NewCounter(memStorage, logger)
	gaugeLogic := logic.New(memStorage, logger)

	updateMetricPathParamsHandler := middleware.
		NewBuilder(handlers.NewUpdateMetricPathParams(counterLogic, gaugeLogic, logger)).
		WithLogger(logger).
		WithRequestDecompression(logger).
		Build()

	updateMetricHandler := middleware.
		NewBuilder(handlers.NewUpdateMetric(counterLogic, gaugeLogic, logger)).
		WithLogger(logger).
		WithRequestDecompression(logger).
		Build()

	getMetricValuePathParamsHandler := middleware.
		NewBuilder(handlers.NewGetMetricValuePathParams(memStorage, memStorage, logger)).
		WithLogger(logger).
		WithRequestDecompression(logger).
		WithResponseCompression(logger).
		Build()

	getMetricValueHandler := middleware.
		NewBuilder(handlers.NewGetMetricValue(memStorage, memStorage, logger)).
		WithLogger(logger).
		WithRequestDecompression(logger).
		WithResponseCompression(logger).
		Build()

	getAllMetricsHandler := middleware.
		NewBuilder(handlers.NewGetAllMetrics(memStorage, logger)).
		WithLogger(logger).
		WithRequestDecompression(logger).
		WithResponseCompression(logger).
		Build()

	pingHandler := middleware.
		NewBuilder(handlers.NewPing(db, logger)).
		WithLogger(logger).
		Build()

	router := chi.NewRouter()

	router.Post(protocol.UpdateMetricURL, updateMetricHandler.ServeHTTP)
	router.Post(protocol.UpdateMetricPathParamsURL, updateMetricPathParamsHandler.ServeHTTP)
	router.Post(protocol.GetMetricURL, getMetricValueHandler.ServeHTTP)
	router.Get(protocol.GetMetricPathParamsURL, getMetricValuePathParamsHandler.ServeHTTP)
	router.Get(protocol.GetAllMetricsURL, getAllMetricsHandler.ServeHTTP)
	router.Get(protocol.PingURL, pingHandler.ServeHTTP)

	return router
}
