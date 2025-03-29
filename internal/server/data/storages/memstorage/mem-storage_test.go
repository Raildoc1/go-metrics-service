package memstorage

import (
	"fmt"
	"math"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
)

func BenchmarkMemStorage_Set(b *testing.B) {
	memStorage := New(zap.NewNop())
	b.ResetTimer()
	for i := range b.N {
		memStorage.Set("test_key", i)
	}
}

func BenchmarkMemStorage_Get(b *testing.B) {
	memStorage := New(zap.NewNop())
	memStorage.Set("test_key", "test_value")
	b.ResetTimer()
	for range b.N {
		_, _ = memStorage.Get("test_key")
	}
}

func BenchmarkMemStorage_GetAll(b *testing.B) {
	memStorage := New(zap.NewNop())
	for i := range 10000 {
		memStorage.Set(fmt.Sprintf("test_key_%v", i), i)
	}
	b.ResetTimer()
	for range b.N {
		_ = memStorage.GetAll()
	}
}

func TestGetExistingValue(t *testing.T) {
	memStorage := New(zap.NewNop())
	tests := []struct {
		name  string
		value any
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
			key := fmt.Sprintf("test-%d", i)
			memStorage.Set(key, test.value)
			val, ok := memStorage.Get(key)
			assert.True(t, ok)
			if ok {
				assert.Equal(t, test.value, val)
			}
		})
	}
}

func TestGetNonExistingValue(t *testing.T) {
	memStorage := New(zap.NewNop())
	_, ok := memStorage.Get("non_existing_key")
	assert.False(t, ok)
}
