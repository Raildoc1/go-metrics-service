package repositories

import (
	"github.com/stretchr/testify/require"
	"go-metrics-service/internal/server/data/storage/memory"
	"testing"
)

func TestKeysCollision(t *testing.T) {
	ms := memory.NewMemStorage()

	cRep := NewCounterRepository(ms)
	gRep := NewGaugeRepository(ms)

	require.NoError(t, cRep.Set("test_counter", 3))

	_, err := gRep.Get("test_counter")
	require.Error(t, err)
	require.ErrorIs(t, err, WrongTypeError)

	err = gRep.Set("test_counter", 3.5)
	require.Error(t, err)
	require.ErrorIs(t, err, WrongTypeError)
}
