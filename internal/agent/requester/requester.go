package requester

import (
	"encoding/json"
	"fmt"
	"go-metrics-service/internal/common/protocol"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type Requester struct {
	host string
}

func New(host string) *Requester {
	return &Requester{
		host: host,
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
	url := "http://" + r.host + protocol.UpdateJsonURL

	body, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	resp, err := resty.New().
		R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(url)

	if err != nil {
		return nil, fmt.Errorf("%w: update failed", err)
	}

	return resp, nil
}
