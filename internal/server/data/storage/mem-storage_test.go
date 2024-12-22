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
			ms := NewMemStorage(zap.NewNop())

			ms.Set("key", test.value)
			val, ok := ms.Get("key")

			require.True(t, ok)
			assert.Equal(t, test.value, val)
		})
	}
}

func TestGetNonExistingValue(t *testing.T) {
	ms := NewMemStorage(zap.NewNop())
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

	const filePath = "test.gz"

	defer func() {
		_ = os.Remove(filePath)
	}()

	err := createAndSaveStorage(testData, filePath)
	require.NoError(t, err)

	ms, err := loadFromFile(filePath)
	require.NoError(t, err)

	for k, v := range testData {
		val, ok := ms.Get(k)
		require.Equal(t, true, ok)
		assert.Equal(t, v, val)
		assert.Equal(t, reflect.TypeOf(v), reflect.TypeOf(val))
	}
}

func createAndSaveStorage(data map[string]interface{}, filePath string) error {
	ms := NewMemStorage(zap.NewNop())
	for k, v := range data {
		ms.Set(k, v)
	}
	return SaveMemStorageToFile(ms, filePath, zap.NewNop())
}

//nolint:wrapcheck // wrapping unnecessary
func loadFromFile(filePath string) (*MemStorage, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	ms, err := LoadFrom(file, zap.NewNop())
	if err != nil {
		return nil, err
	}
	return ms, nil
}
