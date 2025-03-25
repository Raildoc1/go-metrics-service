package testutils

import (
	"encoding/json"
	"go-metrics-service/internal/common/protocol"
	"testing"

	"github.com/stretchr/testify/require"
)

func TCreateCounterDeltaJSON(t *testing.T, id string, delta int64) string {
	t.Helper()
	obj := createCounter(id, delta)
	res, err := json.Marshal(obj)
	require.NoError(t, err)
	return string(res)
}

func TCreateGaugeJSON(t *testing.T, id string, value float64) string {
	t.Helper()
	obj := createGauge(id, value)
	res, err := json.Marshal(obj)
	require.NoError(t, err)
	return string(res)
}

func BCreateCounterDeltaJSON(b *testing.B, id string, delta int64) string {
	b.Helper()
	obj := createCounter(id, delta)
	res, err := json.Marshal(obj)
	if err != nil {
		b.Fatal(err)
	}
	return string(res)
}

func createCounter(id string, delta int64) protocol.Metrics {
	deltaCopy := delta
	return protocol.Metrics{
		ID:    id,
		MType: protocol.Counter,
		Delta: &deltaCopy,
	}
}

func createGauge(id string, value float64) protocol.Metrics {
	valueCopy := value
	return protocol.Metrics{
		ID:    id,
		MType: protocol.Gauge,
		Value: &valueCopy,
	}
}
