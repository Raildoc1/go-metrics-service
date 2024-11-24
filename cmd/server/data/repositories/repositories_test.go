package repositories

import (
	"github.com/stretchr/testify/require"
	"go-metrics-service/cmd/server/data/storage"
	"testing"
)

func TestKeysCollision(t *testing.T) {
	ms := storage.NewMemStorage()

	cRep := NewCounterRepository(ms)
	gRep := NewGaugeRepository(ms)

	require.NoError(t, cRep.Set("test_counter", 3))

	var wrongTypeError storage.WrongTypeError

	_, err := gRep.Get("test_counter")
	require.Error(t, err)
	require.ErrorAs(t, err, &wrongTypeError)

	err = gRep.Set("test_counter", 3.5)
	require.Error(t, err)
	require.ErrorAs(t, err, &wrongTypeError)
}
