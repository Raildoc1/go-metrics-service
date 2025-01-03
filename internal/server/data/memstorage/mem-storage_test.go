package memstorage

import (
	"fmt"
	"math"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetExistingCounter(t *testing.T) {
	memStorage := NewMemStorage(zap.NewNop())
	tests := []struct {
		name  string
		value int64
	}{
		{
			name:  "int64 simple",
			value: 123,
		},
		{
			name:  "int64 zero",
			value: 0,
		},
		{
			name:  "int64 negative",
			value: -234,
		},
	}

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := fmt.Sprintf("counter-%d", i)
			err := memStorage.SetCounter(key, test.value)
			require.NoError(t, err)
			has, err := memStorage.Has(key)
			require.NoError(t, err)
			assert.Equal(t, true, has)
			val, err := memStorage.GetCounter(key)
			require.NoError(t, err)
			assert.Equal(t, test.value, val)
		})
	}
}

func TestGetExistingGauge(t *testing.T) {
	memStorage := NewMemStorage(zap.NewNop())
	tests := []struct {
		name  string
		value float64
	}{
		{
			name:  "float64 simple",
			value: 123.34,
		},
		{
			name:  "float64 zero",
			value: 0,
		},
		{
			name:  "float64 negative",
			value: -0.45,
		},
		{
			name:  "float64 negative zero",
			value: math.Copysign(0, -1),
		},
	}

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := fmt.Sprintf("gauge-%d", i)
			err := memStorage.SetGauge(key, test.value)
			require.NoError(t, err)
			has, err := memStorage.Has(key)
			require.NoError(t, err)
			assert.Equal(t, true, has)
			val, err := memStorage.GetGauge(key)
			require.NoError(t, err)
			assert.Equal(t, test.value, val)
		})
	}
}

func TestGetNonExistingValue(t *testing.T) {
	memStorage := NewMemStorage(zap.NewNop())
	_, err := memStorage.GetCounter("non_existing_key")
	require.Error(t, err)
	_, err = memStorage.GetGauge("non_existing_key")
	require.Error(t, err)
}
