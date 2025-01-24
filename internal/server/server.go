package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/controllers"
	"go-metrics-service/internal/server/handlers"
	"go-metrics-service/internal/server/logic"
	"go-metrics-service/internal/server/middleware"
	"hash"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Repository interface {
	handlers.GaugeRepository
	handlers.CounterRepository
	handlers.AllMetricsRepository
	logic.Repository
}

type TransactionManager interface {
	controllers.TransactionManager
}

type Server struct {
	logger     *zap.Logger
	httpServer *http.Server
	cfg        Config
}

func New(
	cfg Config,
	repository Repository,
	transactionManager TransactionManager,
	pingables []handlers.Pingable,
	logger *zap.Logger,
) *Server {
	var h hash.Hash = nil
	if cfg.SHA256Key != "" {
		h = hmac.New(sha256.New, []byte(cfg.SHA256Key))
	}

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: createMux(h, repository, transactionManager, pingables, logger),
	}

	res := &Server{
		cfg:        cfg,
		logger:     logger,
		httpServer: srv,
	}

	return res
}

func (s *Server) Run() error {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server ListenAndServe failed: %w", err)
	}
	return nil
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	return nil
}

func createMux(
	h hash.Hash,
	repository Repository,
	transactionManager TransactionManager,
	pingables []handlers.Pingable,
	logger *zap.Logger,
) *chi.Mux {
	service := logic.NewService(repository, logger)
	controller := controllers.NewController(transactionManager, service, logger)

	updateMetricPathParamsHandler := middleware.
		NewBuilder(handlers.NewUpdateMetricPathParams(controller, logger)).
		WithLogger(logger).
		WithResponseHash(h, logger).
		WithRequestDecompression(logger).
		WithHashValidation(h, logger).
		Build()

	updateMetricHandler := middleware.
		NewBuilder(handlers.NewUpdateMetric(controller, logger)).
		WithLogger(logger).
		WithResponseHash(h, logger).
		WithRequestDecompression(logger).
		WithHashValidation(h, logger).
		Build()

	updateMetricsHandler := middleware.
		NewBuilder(handlers.NewUpdateMetrics(controller, logger)).
		WithLogger(logger).
		WithResponseHash(h, logger).
		WithRequestDecompression(logger).
		WithHashValidation(h, logger).
		Build()

	getMetricValuePathParamsHandler := middleware.
		NewBuilder(handlers.NewGetMetricValuePathParams(repository, repository, logger)).
		WithLogger(logger).
		WithResponseHash(h, logger).
		WithRequestDecompression(logger).
		WithResponseCompression(logger).
		Build()

	getMetricValueHandler := middleware.
		NewBuilder(handlers.NewGetMetricValue(repository, repository, logger)).
		WithLogger(logger).
		WithResponseHash(h, logger).
		WithRequestDecompression(logger).
		WithResponseCompression(logger).
		Build()

	getAllMetricsHandler := middleware.
		NewBuilder(handlers.NewGetAllMetrics(repository, logger)).
		WithLogger(logger).
		WithResponseHash(h, logger).
		WithRequestDecompression(logger).
		WithResponseCompression(logger).
		Build()

	pingHandler := middleware.
		NewBuilder(handlers.NewPing(pingables, logger)).
		WithLogger(logger).
		Build()

	router := chi.NewRouter()

	router.Post(protocol.UpdateMetricURL, updateMetricHandler.ServeHTTP)
	router.Post(protocol.UpdateMetricsURL, updateMetricsHandler.ServeHTTP)
	router.Post(protocol.UpdateMetricPathParamsURL, updateMetricPathParamsHandler.ServeHTTP)
	router.Post(protocol.GetMetricURL, getMetricValueHandler.ServeHTTP)
	router.Get(protocol.GetMetricPathParamsURL, getMetricValuePathParamsHandler.ServeHTTP)
	router.Get(protocol.GetAllMetricsURL, getAllMetricsHandler.ServeHTTP)
	router.Get(protocol.PingURL, pingHandler.ServeHTTP)

	return router
}
