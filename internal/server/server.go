package server

import (
	"context"
	"errors"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data/repository"
	"go-metrics-service/internal/server/data/storage"
	"go-metrics-service/internal/server/handlers"
	"go-metrics-service/internal/server/logic"
	"go-metrics-service/internal/server/middleware"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func Run(cfg Config, logger *zap.Logger) {
	memStorage := storage.NewMemStorage(logger)
	if cfg.NeedRestore {
		if _, err := os.Stat(cfg.FilePath); err == nil {
			err := memStorage.LoadFromFile(cfg.FilePath)
			if err != nil {
				logger.Error("failed to load from file", zap.Error(err))
			} else {
				logger.Info("data successfully restored", zap.String("path", cfg.FilePath))
			}
		}
	}

	srv := &http.Server{Addr: cfg.ServerAddress}
	srv.Handler = createMux(memStorage, logger)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to start server", zap.Error(err))
		}
	}()

	lifecycle(cfg, logger, memStorage)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("failed to gracefully shutdown", zap.Error(err))
	}
}

func lifecycle(cfg Config, logger *zap.Logger, memStorage *storage.MemStorage) {
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
			trySaveStorage(cfg.FilePath, logger, memStorage)
		case <-cancelChan:
			trySaveStorage(cfg.FilePath, logger, memStorage)
			return
		}
	}
}

func trySaveStorage(filePath string, logger *zap.Logger, memStorage *storage.MemStorage) {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		const dirPerm = 0o700
		err = os.MkdirAll(dir, dirPerm)
		if err != nil {
			logger.Error("failed to create directory", zap.Error(err))
			return
		}
	}
	if err := memStorage.SaveToFile(filePath); err != nil {
		logger.Error("failed to save to file", zap.Error(err))
	} else {
		logger.Info("successfully saved to file", zap.String("path", filePath))
	}
}

func createMux(strg repository.Storage, logger *zap.Logger) *chi.Mux {
	rep := repository.New(strg)

	counterLogic := logic.NewCounter(rep, logger)
	gaugeLogic := logic.New(rep, logger)

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
		NewBuilder(handlers.NewGetMetricValuePathParams(rep, rep, logger)).
		WithLogger(logger).
		WithRequestDecompression(logger).
		WithResponseCompression(logger).
		Build()

	getMetricValueHandler := middleware.
		NewBuilder(handlers.NewGetMetricValue(rep, rep, logger)).
		WithLogger(logger).
		WithRequestDecompression(logger).
		WithResponseCompression(logger).
		Build()

	getAllMetricsHandler := middleware.
		NewBuilder(handlers.NewGetAllMetrics(strg, logger)).
		WithLogger(logger).
		WithRequestDecompression(logger).
		WithResponseCompression(logger).
		Build()

	router := chi.NewRouter()

	router.Post(protocol.UpdateMetricURL, updateMetricHandler.ServeHTTP)
	router.Post(protocol.UpdateMetricPathParamsURL, updateMetricPathParamsHandler.ServeHTTP)
	router.Post(protocol.GetMetricURL, getMetricValueHandler.ServeHTTP)
	router.Get(protocol.GetMetricPathParamsURL, getMetricValuePathParamsHandler.ServeHTTP)
	router.Get(protocol.GetAllMetricsURL, getAllMetricsHandler.ServeHTTP)

	return router
}
