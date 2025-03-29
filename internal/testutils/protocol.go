package testutils

import (
	"encoding/json"
	"go-metrics-service/internal/common/protocol"
	"testing"

	"github.com/stretchr/testify/require"
)

func TCreateCounterDeltaJSON(t *testing.T, id string, delta int64) string {
	t.Helper()
	obj := CreateCounter(id, delta)
	res, err := json.Marshal(obj)
	require.NoError(t, err)
	return string(res)
}

func TCreateGaugeDiffJSON(t *testing.T, id string, value float64) string {
	t.Helper()
	obj := CreateGauge(id, value)
	res, err := json.Marshal(obj)
	require.NoError(t, err)
	return string(res)
}

func BCreateCounterDeltaJSON(b *testing.B, id string, delta int64) string {
	b.Helper()
	obj := CreateCounter(id, delta)
	res, err := json.Marshal(obj)
	if err != nil {
		b.Fatal(err)
	}
	return string(res)
}

func CreateCounter(id string, delta int64) protocol.Metrics {
	deltaCopy := delta
	return protocol.Metrics{
		ID:    id,
		MType: protocol.Counter,
		Delta: &deltaCopy,
	}
}

func CreateGauge(id string, value float64) protocol.Metrics {
	valueCopy := value
	return protocol.Metrics{
		ID:    id,
		MType: protocol.Gauge,
		Value: &valueCopy,
	}
}

func TCreateMetricsJSON(t *testing.T, values []protocol.Metrics) string {
	t.Helper()
	res, err := json.Marshal(values)
	require.NoError(t, err)
	return string(res)
}

func BCreateMetricsJSON(b *testing.B, values []protocol.Metrics) string {
	b.Helper()
	res, err := json.Marshal(values)
	require.NoError(b, err)
	return string(res)
}
