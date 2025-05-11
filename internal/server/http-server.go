// Package server contains composition root for server
package server

import (
	"context"
	"errors"
	"fmt"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/handlers"
	"go-metrics-service/internal/server/logic"
	"go-metrics-service/internal/server/middleware"
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

type Controller interface {
	handlers.MetricController
}

type HTTPServer struct {
	logger     *zap.Logger
	httpServer *http.Server
	controller Controller
	cfg        Config
}

type middlewareFactory interface {
	CreateHandler(next http.Handler) http.Handler
}

func NewHTTP(
	cfg Config,
	repository Repository,
	hashFactory middleware.HashFactory,
	pingables []handlers.Pingable,
	logger *zap.Logger,
	decoder middleware.Decoder,
	controller Controller,
) (*HTTPServer, error) {
	mux, err := createMux(
		hashFactory,
		repository,
		controller,
		pingables,
		logger,
		decoder,
		cfg.TrustedSubnet,
	)

	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: mux,
	}

	res := &HTTPServer{
		cfg:        cfg,
		controller: controller,
		logger:     logger,
		httpServer: srv,
	}

	return res, nil
}

func (s *HTTPServer) Run() error {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server ListenAndServe failed: %w", err)
	}
	return nil
}

func (s *HTTPServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	return nil
}

func createMux(
	hashFactory middleware.HashFactory,
	repository Repository,
	controller Controller,
	pingables []handlers.Pingable,
	logger *zap.Logger,
	decoder middleware.Decoder,
	trustedSubnet string,
) (*chi.Mux, error) {
	loggerMiddleware := middleware.NewLogger(logger)

	var subnetFilterMiddleware middlewareFactory

	if trustedSubnet != "" {
		mw, err := middleware.NewSubnetFilter(logger, trustedSubnet)
		if err != nil {
			return nil, fmt.Errorf("failed to create subnet filter: %w", err)
		}
		subnetFilterMiddleware = mw
	} else {
		subnetFilterMiddleware = middleware.NewNop()
	}

	var decryptMiddleware middlewareFactory

	if decoder != nil {
		decryptMiddleware = middleware.NewRequestDecoder(decoder, logger)
	} else {
		decryptMiddleware = middleware.NewNop()
	}

	var responseHashMiddleware middlewareFactory
	var requestHashMiddleware middlewareFactory

	if hashFactory != nil {
		responseHashMiddleware = middleware.NewResponseHash(logger, hashFactory)
		requestHashMiddleware = middleware.NewRequestHash(logger, hashFactory)
	} else {
		responseHashMiddleware = middleware.NewNop()
		requestHashMiddleware = middleware.NewNop()
	}

	requestDecompressMiddleware := middleware.NewRequestDecompressor(logger)
	responseCompressMiddleware := middleware.NewResponseCompressor(logger)

	updateMetricPathParamsHandler := handlers.NewUpdateMetricPathParams(controller, logger)
	updateMetricHandler := handlers.NewUpdateMetric(controller, logger)
	updateMetricsHandler := handlers.NewUpdateMetrics(controller, logger)
	getMetricValuePathParamsHandler := handlers.NewGetMetricValuePathParams(repository, repository, logger)
	getMetricValueHandler := handlers.NewGetMetricValue(repository, repository, logger)
	getAllMetricsHandler := handlers.NewGetAllMetrics(repository, logger)
	pingHandler := handlers.NewPing(pingables, logger)

	router := chi.NewRouter()

	router.With(
		loggerMiddleware.CreateHandler,
		subnetFilterMiddleware.CreateHandler,
		decryptMiddleware.CreateHandler,
		requestHashMiddleware.CreateHandler,
		responseHashMiddleware.CreateHandler,
		requestDecompressMiddleware.CreateHandler,
	).Route("/", func(router chi.Router) {
		router.Post(protocol.UpdateMetricURL, updateMetricHandler.ServeHTTP)
		router.Post(protocol.UpdateMetricsURL, updateMetricsHandler.ServeHTTP)
		router.Post(protocol.UpdateMetricPathParamsURL, updateMetricPathParamsHandler.ServeHTTP)
		router.Get(protocol.PingURL, pingHandler.ServeHTTP)
		router.With(responseCompressMiddleware.CreateHandler).
			Route("/", func(router chi.Router) {
				router.Post(protocol.GetMetricURL, getMetricValueHandler.ServeHTTP)
				router.Get(protocol.GetMetricPathParamsURL, getMetricValuePathParamsHandler.ServeHTTP)
				router.Get(protocol.GetAllMetricsURL, getAllMetricsHandler.ServeHTTP)
			})
	})

	return router, nil
}
