package server

import (
	"context"
	"errors"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data/repository"
	"go-metrics-service/internal/server/data/storage"
	"go-metrics-service/internal/server/handlers"
	"go-metrics-service/internal/server/logic/counter"
	"go-metrics-service/internal/server/logic/gauge"
	"go-metrics-service/internal/server/middleware"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

type Logger interface {
	middleware.Logger
	handlers.Logger
}

func Run(cfg Config, logger Logger) {
	memStorage := storage.NewMemStorage(logger)
	if cfg.NeedRestore {
		if _, err := os.Stat(cfg.FilePath); err == nil {
			err := memStorage.LoadFromFile(cfg.FilePath)
			if err != nil {
				logger.Errorln(err)
			} else {
				logger.Infoln("Data successfully restored")
			}
		}
	}

	srv := &http.Server{Addr: cfg.ServerAddress}
	srv.Handler = createMux(memStorage, logger)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorln(err)
		}
	}()

	lifecycle(cfg, logger, memStorage)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorln(err)
	}
}

func lifecycle(cfg Config, logger Logger, memStorage *storage.MemStorage) {
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

func trySaveStorage(filePath string, logger Logger, memStorage *storage.MemStorage) {
	if err := memStorage.SaveToFile(filePath); err != nil {
		logger.Errorln(err)
	} else {
		logger.Infoln("Data successfully saved")
	}
}

func createMux(storage repository.Storage, logger Logger) *chi.Mux {
	rep := repository.New(storage)

	counterLogic := counter.New(rep)
	gaugeLogic := gauge.New(rep)

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
		NewBuilder(handlers.NewGetAllMetrics(storage, logger)).
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
