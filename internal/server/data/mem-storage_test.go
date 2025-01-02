package data

import (
	"os"
	"reflect"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage(t *testing.T) {
	memStorage := NewMemStorage(zap.NewNop())
	t.Run("Get existing counter", func(t *testing.T) {
		testGetExistingCounter(memStorage, t)
	})
	t.Run("Get existing gauge", func(t *testing.T) {
		testGetExistingGauge(memStorage, t)
	})
	t.Run("Get non existing value", func(t *testing.T) {
		testGetNonExistingValue(memStorage, t)
	})
}

func TestFileSaveLoad(t *testing.T) {
	counters := map[string]int64{
		"test_key_1": int64(3),
		"test_key_2": int64(-34),
		"test_key_3": int64(234),
	}
	gauges := map[string]float64{
		"test_key_4": float64(34.6),
		"test_key_5": float64(-43.23),
	}

	const filePath = "test.gz"

	defer func() {
		_ = os.Remove(filePath)
	}()

	memStorageToSave := NewMemStorage(zap.NewNop())
	for k, v := range counters {
		err := memStorageToSave.SetCounter(k, v)
		if err != nil {
			require.NoError(t, err)
		}
	}
	for k, v := range gauges {
		err := memStorageToSave.SetGauge(k, v)
		if err != nil {
			require.NoError(t, err)
		}
	}
	err := SaveMemStorageToFile(memStorageToSave, filePath, zap.NewNop())
	require.NoError(t, err)

	loadedMemStorage, err := loadFromFile(filePath)
	require.NoError(t, err)

	for k, v := range counters {
		val, err := loadedMemStorage.GetCounter(k)
		require.NoError(t, err)
		assert.Equal(t, v, val)
		assert.Equal(t, reflect.TypeOf(v), reflect.TypeOf(val))
	}

	for k, v := range gauges {
		val, err := loadedMemStorage.GetGauge(k)
		require.NoError(t, err)
		assert.Equal(t, v, val)
		assert.Equal(t, reflect.TypeOf(v), reflect.TypeOf(val))
	}
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
