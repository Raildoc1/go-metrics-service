package backupmemstorage

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

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

	toSave := newEmpty(zap.NewNop())
	for k, v := range counters {
		err := toSave.SetCounter(k, v)
		if err != nil {
			require.NoError(t, err)
		}
	}
	for k, v := range gauges {
		err := toSave.SetGauge(k, v)
		if err != nil {
			require.NoError(t, err)
		}
	}
	err := toSave.SaveToFile(filePath)
	require.NoError(t, err)

	loaded, err := loadFromFile(filePath, zap.NewNop())
	require.NoError(t, err)

	for k, v := range counters {
		val, err := loaded.GetCounter(k)
		require.NoError(t, err)
		assert.Equal(t, v, val)
		assert.Equal(t, reflect.TypeOf(v), reflect.TypeOf(val))
	}

	for k, v := range gauges {
		val, err := loaded.GetGauge(k)
		require.NoError(t, err)
		assert.Equal(t, v, val)
		assert.Equal(t, reflect.TypeOf(v), reflect.TypeOf(val))
	}
}
