package storage

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetExistingValue(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{
			name:  "float64 simple",
			value: float64(123),
		},
		{
			name:  "float64 zero",
			value: float64(0),
		},
		{
			name:  "float64 negative zero",
			value: float64(-0),
		},
		{
			name:  "int64 simple",
			value: int64(123),
		},
		{
			name:  "int64 zero",
			value: int64(0),
		},
		{
			name:  "int64 negative",
			value: int64(-234),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms := NewMemStorage()

			ms.Set("key", test.value)
			val, ok := ms.Get("key")

			require.True(t, ok)
			assert.Equal(t, test.value, val)
		})
	}
}

func TestGetNonExistingValue(t *testing.T) {
	ms := NewMemStorage()
	_, ok := ms.Get("non_existing_key")
	require.False(t, ok)
}
