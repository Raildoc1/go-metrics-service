package repository

import (
	"go-metrics-service/internal/server/data/storage/memory"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeysCollision(t *testing.T) {
	ms := memory.NewMemStorage()

	rep := New(ms)

	require.NoError(t, rep.SetInt64("test_counter", 3))

	_, err := rep.GetFloat64("test_counter")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrWrongType)

	err = rep.SetFloat64("test_counter", 3.5)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrWrongType)
}
