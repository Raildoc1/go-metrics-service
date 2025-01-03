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
	storage    Storage
	pingables  []handlers.Pingable
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
		storage:    storage,
		pingables:  pingables,
		logger:     logger,
		httpServer: srv,
	}

	srv.Handler = res.createMux()

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

func (s *Server) createMux() *chi.Mux {
	counterLogic := logic.NewCounter(s.storage, s.logger)
	gaugeLogic := logic.New(s.storage, s.logger)

	updateMetricPathParamsHandler := middleware.
		NewBuilder(handlers.NewUpdateMetricPathParams(counterLogic, gaugeLogic, s.logger)).
		WithLogger(s.logger).
		WithRequestDecompression(s.logger).
		Build()

	updateMetricHandler := middleware.
		NewBuilder(handlers.NewUpdateMetric(counterLogic, gaugeLogic, s.logger)).
		WithLogger(s.logger).
		WithRequestDecompression(s.logger).
		Build()

	getMetricValuePathParamsHandler := middleware.
		NewBuilder(handlers.NewGetMetricValuePathParams(s.storage, s.storage, s.logger)).
		WithLogger(s.logger).
		WithRequestDecompression(s.logger).
		WithResponseCompression(s.logger).
		Build()

	getMetricValueHandler := middleware.
		NewBuilder(handlers.NewGetMetricValue(s.storage, s.storage, s.logger)).
		WithLogger(s.logger).
		WithRequestDecompression(s.logger).
		WithResponseCompression(s.logger).
		Build()

	getAllMetricsHandler := middleware.
		NewBuilder(handlers.NewGetAllMetrics(s.storage, s.logger)).
		WithLogger(s.logger).
		WithRequestDecompression(s.logger).
		WithResponseCompression(s.logger).
		Build()

	pingHandler := middleware.
		NewBuilder(handlers.NewPing(s.pingables, s.logger)).
		WithLogger(s.logger).
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
