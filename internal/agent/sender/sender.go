package sender

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	storagePkg "go-metrics-service/internal/agent/storage"
	"go-metrics-service/internal/common/compression"
	"go-metrics-service/internal/common/protocol"
	"io"
	"net/http"
	"syscall"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

var (
	ErrServerUnavailable = errors.New("server unavailable")
)

type Sender struct {
	logger  *zap.Logger
	storage *storagePkg.Storage
	host    string
}

func New(
	host string,
	storage *storagePkg.Storage,
	logger *zap.Logger,
) *Sender {
	return &Sender{
		host:    host,
		storage: storage,
		logger:  logger,
	}
}

func (s *Sender) Send() error {
	metricsDiff := s.storage.GetUncommitedData()
	metricsToUpdateCount := len(metricsDiff.CounterDeltas) + len(metricsDiff.GaugeValues)
	if metricsToUpdateCount == 0 {
		return nil
	}
	metricsToSend := make([]protocol.Metrics, 0, metricsToUpdateCount)
	for k, v := range metricsDiff.CounterDeltas {
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
	for k, v := range metricsDiff.GaugeValues {
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
	err := s.sendUpdates(metricsToSend)
	if err != nil {
		return fmt.Errorf("sending updates failed: %w", err)
	}
	s.storage.Commit()
	return nil
}
func (s *Sender) sendUpdates(metrics []protocol.Metrics) error {
	url := s.createURL(protocol.UpdateMetricsURL)
	var body bytes.Buffer
	err := compression.GzipCompress(
		metrics,
		func(writer io.Writer) compression.Encoder {
			return json.NewEncoder(writer)
		},
		&body,
		gzip.BestSpeed,
		s.logger,
	)
	if err != nil {
		return fmt.Errorf("failed to compress request: %w", err)
	}
	resp, err := resty.New().
		R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body.Bytes()).
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
