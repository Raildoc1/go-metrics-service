package server

import (
	"context"
	"errors"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/handlers"
	"go-metrics-service/internal/server/logic"
	"go-metrics-service/internal/server/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Storage interface {
	handlers.GaugeRepository
	handlers.CounterRepository
	handlers.AllMetricsRepository
	logic.CounterRepository
	logic.GaugeRepository
}

type Server struct {
	cfg        Config
	logger     *zap.Logger
	httpServer *http.Server
}

func New(
	cfg Config,
	storage Storage,
	pingables []handlers.Pingable,
	logger *zap.Logger,
) *Server {
	srv := &http.Server{Addr: cfg.ServerAddress}

	res := &Server{
		cfg:        cfg,
		logger:     logger,
		httpServer: srv,
	}

	srv.Handler = createMux(storage, pingables, logger)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to start server", zap.Error(err))
		}
	}()

	return res
}

func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("failed to gracefully shutdown", zap.Error(err))
	}
}

func createMux(
	storage Storage,
	pingables []handlers.Pingable,
	logger *zap.Logger,
) *chi.Mux {
	counterLogic := logic.NewCounter(storage, logger)
	gaugeLogic := logic.New(storage, logger)

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
		NewBuilder(handlers.NewGetMetricValuePathParams(storage, storage, logger)).
		WithLogger(logger).
		WithRequestDecompression(logger).
		WithResponseCompression(logger).
		Build()

	getMetricValueHandler := middleware.
		NewBuilder(handlers.NewGetMetricValue(storage, storage, logger)).
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

	pingHandler := middleware.
		NewBuilder(handlers.NewPing(pingables, logger)).
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
