package backupmemstorage

import (
	"os"
	"reflect"
	"testing"
	"time"

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

	backupConfig := BackupConfig{
		FilePath:      filePath,
		StoreInterval: time.Second * 1000,
	}

	defer func() {
		_ = os.Remove(filePath)
	}()

	toSave := newEmpty(backupConfig, zap.NewNop())
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
	err := toSave.saveToFile(filePath)
	require.NoError(t, err)

	loaded, err := loadFromFile(backupConfig, zap.NewNop())
	require.NoError(t, err)
	defer loaded.Stop()

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
