package repository

import (
	"go-metrics-service/internal/server/data"
	"go-metrics-service/internal/server/data/storage"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/require"
)

func TestKeysCollision(t *testing.T) {
	ms := storage.NewMemStorage(zap.NewNop())

	rep := New(ms)

	require.NoError(t, rep.SetInt64("test_counter", 3))

	_, err := rep.GetFloat64("test_counter")
	require.Error(t, err)
	require.ErrorIs(t, err, data.ErrWrongType)

	err = rep.SetFloat64("test_counter", 3.5)
	require.Error(t, err)
	require.ErrorIs(t, err, data.ErrWrongType)
}
