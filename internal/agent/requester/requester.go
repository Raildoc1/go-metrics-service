package requester

import (
	"fmt"
	"go-metrics-service/internal/common/protocol"
	"net/http"
	"strconv"

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
	resp, err := r.sendUpdate(protocol.Counter, metricName, strconv.FormatInt(delta, 10))
	if err != nil {
		return fmt.Errorf("failed to send counter delta to %s: %w", r.host, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to send counter delta to %s: code %d", r.host, resp.StatusCode())
	}
	return nil
}

func (r *Requester) SendGauge(metricName string, value float64) error {
	resp, err := r.sendUpdate(protocol.Gauge, metricName, strconv.FormatFloat(value, 'f', -1, 64))
	if err != nil {
		return fmt.Errorf("failed to send gauge to %s: %w", r.host, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to send gauge to %s: code %d", r.host, resp.StatusCode())
	}
	return nil
}

func (r *Requester) sendUpdate(metricType string, metricKey string, metricValue string) (*resty.Response, error) {
	url := "http://" + r.host + protocol.UpdateMetricValueURL

	resp, err := resty.New().
		SetPathParams(
			map[string]string{
				protocol.TypeParam:  metricType,
				protocol.KeyParam:   metricKey,
				protocol.ValueParam: metricValue,
			}).
		R().
		Post(url)

	if err != nil {
		return nil, fmt.Errorf("%w: update failed", err)
	}

	return resp, nil
}
