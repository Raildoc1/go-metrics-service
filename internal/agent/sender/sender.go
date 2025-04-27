package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	storagePkg "go-metrics-service/internal/agent/storage"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/pkg/compression"
	"go-metrics-service/pkg/gohelpers"
	"go-metrics-service/pkg/timeutils"
	"hash"
	"io"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

var (
	ErrServerUnavailable = errors.New("server unavailable")
)

type HashFactory interface {
	Create() hash.Hash
}

type Encoder interface {
	Encode([]byte) ([]byte, error)
}

type Sender struct {
	logger        *zap.Logger
	storage       *storagePkg.Storage
	hashFactory   HashFactory
	doneCh        chan struct{}
	countersCh    chan map[string]int64
	gaugesCh      chan struct{}
	host          string
	attemptsDelay []time.Duration
	encoder       Encoder
	ip            net.IP
}

func New(
	host string,
	attemptsDelay []time.Duration,
	storage *storagePkg.Storage,
	logger *zap.Logger,
	hashFactory HashFactory,
	encoder Encoder,
	ip net.IP,
) *Sender {
	return &Sender{
		host:          host,
		storage:       storage,
		logger:        logger,
		hashFactory:   hashFactory,
		doneCh:        make(chan struct{}),
		countersCh:    make(chan map[string]int64),
		gaugesCh:      make(chan struct{}),
		attemptsDelay: attemptsDelay,
		encoder:       encoder,
		ip:            ip,
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
	url := s.createURL(protocol.UpdateMetricsURL)
	var body bytes.Buffer
	err := compression.GzipCompress(
		metrics,
		func(writer io.Writer) compression.Encoder {
			je := json.NewEncoder(writer)
			je.SetIndent("", "")
			return je
		},
		&body,
		gzip.BestSpeed,
		s.logger,
	)
	if err != nil {
		return fmt.Errorf("failed to compress request: %w", err)
	}

	req := resty.New().
		R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("X-Real-IP", s.ip.String())

	if s.hashFactory != nil {
		h := s.hashFactory.Create()
		_, err = h.Write(body.Bytes())
		if err != nil {
			return fmt.Errorf("failed to hash: %w", err)
		}
		req = req.SetHeader(protocol.HashHeader, hex.EncodeToString(h.Sum(nil)))
	}

	bodyBytes := body.Bytes()

	if s.encoder != nil {
		encoded, err := s.encoder.Encode(bodyBytes)
		if err != nil {
			return fmt.Errorf("failed to encode request: %w", err)
		}
		bodyBytes = encoded
	}

	resp, err := req.
		SetBody(bodyBytes).
		SetLogger(NewRestyLogger(s.logger)).
		SetDebug(true).
		SetContext(ctx).
		Post(url)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			return fmt.Errorf("%w: %w", err, ErrServerUnavailable)
		}
		return fmt.Errorf("%w: update failed", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to send updates to %s: %d", s.host, resp.StatusCode())
	}
	return nil
}

func (s *Sender) createURL(path string) string {
	return fmt.Sprintf("http://%s%s", s.host, path)
}
