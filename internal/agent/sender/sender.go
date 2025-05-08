package sender

import (
	"context"
	storagePkg "go-metrics-service/internal/agent/storage"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/pkg/gohelpers"
	"go-metrics-service/pkg/timeutils"
	"time"

	"go.uber.org/zap"
)

type Driver interface {
	SendUpdates(ctx context.Context, metrics []protocol.Metrics) error
}

type Sender struct {
	logger        *zap.Logger
	storage       *storagePkg.Storage
	doneCh        chan struct{}
	countersCh    chan map[string]int64
	gaugesCh      chan struct{}
	attemptsDelay []time.Duration
	driver        Driver
}

func New(
	attemptsDelay []time.Duration,
	storage *storagePkg.Storage,
	logger *zap.Logger,
	driver Driver,
) *Sender {
	return &Sender{
		storage:       storage,
		logger:        logger,
		doneCh:        make(chan struct{}),
		countersCh:    make(chan map[string]int64),
		gaugesCh:      make(chan struct{}),
		attemptsDelay: attemptsDelay,
		driver:        driver,
	}
}

func (s *Sender) Start(interval time.Duration, workersCount int) chan error {
	errChs := make([]chan error, 0)

	errChs = append(
		errChs,
		gohelpers.StartTickerProcess(s.doneCh, s.Schedule, interval),
		gohelpers.StartProcess[struct{}](
			s.doneCh,
			s.sendGaugesUpdate,
			func() {},
			s.gaugesCh,
		),
	)

	for range workersCount {
		errChs = append(errChs, gohelpers.StartProcess[map[string]int64](
			s.doneCh,
			s.sendCountersUpdate,
			func() {},
			s.countersCh,
		))
	}

	return gohelpers.AggregateErrors(errChs...)
}

func (s *Sender) Stop() {
	close(s.doneCh)
}

func (s *Sender) Schedule(_ context.Context) error {
	s.gaugesCh <- struct{}{}
	c := s.storage.ConsumeUncommitedCounters()
	s.countersCh <- c
	return nil
}

func (s *Sender) sendCountersUpdate(ctx context.Context, counterDeltas map[string]int64) error {
	metricsToSend := make([]protocol.Metrics, 0, len(counterDeltas))

	for k, v := range counterDeltas {
		val := v
		metricsToSend = append(
			metricsToSend,
			protocol.Metrics{
				ID:    k,
				MType: protocol.Counter,
				Delta: &val,
			},
		)
	}

	return s.sendUpdatesWithRetry(ctx, metricsToSend)
}

func (s *Sender) sendGaugesUpdate(ctx context.Context, _ struct{}) error {
	return s.storage.HandleUncommitedGauges( //nolint:wrapcheck // wrapping unnecessary
		func(uncommitedValues map[string]float64) error {
			metricsToSend := make([]protocol.Metrics, 0, len(uncommitedValues))

			for k, v := range uncommitedValues {
				val := v
				metricsToSend = append(
					metricsToSend,
					protocol.Metrics{
						ID:    k,
						MType: protocol.Gauge,
						Value: &val,
					},
				)
			}

			return s.sendUpdates(ctx, metricsToSend)
		})
}

func (s *Sender) sendUpdatesWithRetry(ctx context.Context, metrics []protocol.Metrics) error {
	return timeutils.Retry( //nolint:wrapcheck // wrapping unnecessary
		ctx,
		s.attemptsDelay,
		func(ctx context.Context) error {
			return s.sendUpdates(ctx, metrics)
		},
		func(err error) bool {
			s.logger.Error("sending updates failed", zap.Error(err))
			return true
		})
}

func (s *Sender) sendUpdates(ctx context.Context, metrics []protocol.Metrics) error {
	return s.driver.SendUpdates(ctx, metrics)
}
