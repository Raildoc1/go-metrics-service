package storage

import (
	"math"
	"os"
	"reflect"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			value: math.Copysign(0, -1),
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
			ms := NewMemStorage(zap.NewNop().Sugar())

			ms.Set("key", test.value)
			val, ok := ms.Get("key")

			require.True(t, ok)
			assert.Equal(t, test.value, val)
		})
	}
}

func TestGetNonExistingValue(t *testing.T) {
	ms := NewMemStorage(zap.NewNop().Sugar())
	_, ok := ms.Get("non_existing_key")
	require.False(t, ok)
}

func TestFileSaveLoad(t *testing.T) {
	testData := map[string]interface{}{
		"test_key_1": int64(3),
		"test_key_2": int64(-34),
		"test_key_3": int64(234),
		"test_key_4": float64(34.6),
		"test_key_5": float64(-43.23),
		"test_key_6": "Hello, World!",
		"test_key_7": "",
	}
	ms1 := NewMemStorage(zap.NewNop().Sugar())
	for k, v := range testData {
		ms1.Set(k, v)
	}
	err := ms1.SaveToFile("test.gz")
	defer func() {
		_ = os.Remove("test.gz")
	}()
	require.NoError(t, err)
	ms2 := NewMemStorage(zap.NewNop().Sugar())
	err = ms2.LoadFromFile("test.gz")
	require.NoError(t, err)
	for k, v := range testData {
		val, ok := ms2.Get(k)
		require.Equal(t, true, ok)
		assert.Equal(t, v, val)
		assert.Equal(t, reflect.TypeOf(v), reflect.TypeOf(val))
	}
}
