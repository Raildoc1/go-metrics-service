package requester

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"go-metrics-service/internal/common/compression"
	"go-metrics-service/internal/common/protocol"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/go-resty/resty/v2"
)

type Requester struct {
	logger *zap.Logger
	host   string
}

func New(host string, logger *zap.Logger) *Requester {
	return &Requester{
		host:   host,
		logger: logger,
	}
}

func (r *Requester) SendCounterDelta(metricName string, delta int64) error {
	requestData := protocol.Metrics{
		ID:    metricName,
		MType: protocol.Counter,
		Delta: &delta,
	}
	resp, err := r.sendUpdate(requestData)
	if err != nil {
		return fmt.Errorf("failed to send counter delta to %s: %w", r.host, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to send counter delta to %s: code %d", r.host, resp.StatusCode())
	}
	return nil
}

func (r *Requester) SendGauge(metricName string, value float64) error {
	requestData := protocol.Metrics{
		ID:    metricName,
		MType: protocol.Gauge,
		Value: &value,
	}
	resp, err := r.sendUpdate(requestData)
	if err != nil {
		return fmt.Errorf("failed to send gauge to %s: %w", r.host, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to send gauge to %s: code %d", r.host, resp.StatusCode())
	}
	return nil
}

func (r *Requester) sendUpdate(requestData protocol.Metrics) (*resty.Response, error) {
	url := "http://" + r.host + protocol.UpdateMetricURL

	var body bytes.Buffer

	err := compression.GzipCompress(
		requestData,
		func(writer io.Writer) compression.Encoder {
			return json.NewEncoder(writer)
		},
		&body,
		gzip.BestSpeed,
		r.logger,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to compress request: %w", err)
	}

	resp, err := resty.New().
		R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body.Bytes()).
		Post(url)

	if err != nil {
		return nil, fmt.Errorf("%w: update failed", err)
	}

	return resp, nil
}
