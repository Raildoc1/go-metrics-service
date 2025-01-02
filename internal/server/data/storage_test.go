package data

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type storage interface {
	SetCounter(key string, value int64) error
	SetGauge(key string, value float64) error
	Has(key string) (bool, error)
	GetCounter(key string) (int64, error)
	GetGauge(key string) (float64, error)
	GetAll() (map[string]any, error)
}

func testGetExistingCounter(s storage, t *testing.T) {
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
			err := s.SetCounter(key, test.value)
			require.NoError(t, err)
			has, err := s.Has(key)
			require.NoError(t, err)
			assert.Equal(t, true, has)
			val, err := s.GetCounter(key)
			require.NoError(t, err)
			assert.Equal(t, test.value, val)
		})
	}
}

func testGetExistingGauge(s storage, t *testing.T) {
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
			err := s.SetGauge(key, test.value)
			require.NoError(t, err)
			has, err := s.Has(key)
			require.NoError(t, err)
			assert.Equal(t, true, has)
			val, err := s.GetGauge(key)
			require.NoError(t, err)
			assert.Equal(t, test.value, val)
		})
	}
}

func testGetNonExistingValue(s storage, t *testing.T) {
	_, err := s.GetCounter("non_existing_key")
	require.Error(t, err)
	_, err = s.GetGauge("non_existing_key")
	require.Error(t, err)
}
