package sender

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go-metrics-service/internal/agent/gohelpers"
	storagePkg "go-metrics-service/internal/agent/storage"
	"go-metrics-service/internal/common/compression"
	"go-metrics-service/internal/common/protocol"
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

type Sender struct {
	hash       hash.Hash
	logger     *zap.Logger
	storage    *storagePkg.Storage
	doneCh     chan struct{}
	countersCh chan map[string]int64
	gaugesCh   chan map[string]float64
	host       string
}

func New(
	host string,
	storage *storagePkg.Storage,
	logger *zap.Logger,
	h hash.Hash,
) *Sender {
	return &Sender{
		host:       host,
		storage:    storage,
		logger:     logger,
		hash:       h,
		doneCh:     make(chan struct{}),
		countersCh: make(chan map[string]int64),
		gaugesCh:   make(chan map[string]float64),
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

	return s.sendUpdates(metricsToSend)
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

	if s.hash != nil {
		s.hash.Reset()
		_, err := s.hash.Write(body.Bytes())
		if err != nil {
			return fmt.Errorf("failed to hash: %w", err)
		}
		req = req.SetHeader(protocol.HashHeader, hex.EncodeToString(s.hash.Sum(nil)))
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
