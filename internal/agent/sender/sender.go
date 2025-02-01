package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go-metrics-service/internal/agent/gohelpers"
	storagePkg "go-metrics-service/internal/agent/storage"
	"go-metrics-service/internal/common/compression"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/common/timeutils"
	"hash"
	"io"
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

type Sender struct {
	logger        *zap.Logger
	storage       *storagePkg.Storage
	hashFactory   HashFactory
	doneCh        chan struct{}
	countersCh    chan map[string]int64
	gaugesCh      chan map[string]float64
	host          string
	attemptsDelay []time.Duration
}

func New(
	host string,
	attemptsDelay []time.Duration,
	storage *storagePkg.Storage,
	logger *zap.Logger,
	hashFactory HashFactory,
) *Sender {
	return &Sender{
		host:          host,
		storage:       storage,
		logger:        logger,
		hashFactory:   hashFactory,
		doneCh:        make(chan struct{}),
		countersCh:    make(chan map[string]int64),
		gaugesCh:      make(chan map[string]float64),
		attemptsDelay: attemptsDelay,
	}
}

func (s *Sender) Start(interval time.Duration, workersCount int) chan error {
	errChs := make([]chan error, 0)

	errChs = append(
		errChs,
		gohelpers.StartTickerProcess(s.doneCh, s.Schedule, interval),
		gohelpers.StartProcess[map[string]float64](
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

func (s *Sender) Schedule() error {
	g := s.storage.ConsumeUncommitedGauges()
	s.gaugesCh <- g
	c := s.storage.ConsumeUncommitedCounters()
	s.countersCh <- c
	return nil
}

func (s *Sender) sendCountersUpdate(counterDeltas map[string]int64) error {
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

	return timeutils.Retry(
		context.Background(),
		s.attemptsDelay,
		func(ctx context.Context) error {
			return s.sendUpdates(metricsToSend)
		},
		func(err error) bool {
			return true
		})
}

func (s *Sender) sendGaugesUpdate(gaugeValues map[string]float64) error {
	metricsToSend := make([]protocol.Metrics, 0, len(gaugeValues))

	for k, v := range gaugeValues {
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

	return s.sendUpdates(metricsToSend)
}

func (s *Sender) sendUpdates(metrics []protocol.Metrics) error {
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
		SetHeader("Content-Encoding", "gzip")

	if s.hashFactory != nil {
		h := s.hashFactory.Create()
		_, err := h.Write(body.Bytes())
		if err != nil {
			return fmt.Errorf("failed to hash: %w", err)
		}
		req = req.SetHeader(protocol.HashHeader, hex.EncodeToString(h.Sum(nil)))
	}

	resp, err := req.
		SetBody(body.Bytes()).
		SetLogger(NewRestyLogger(s.logger)).
		SetDebug(true).
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
