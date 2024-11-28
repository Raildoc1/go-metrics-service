package repositories

import (
	"go-metrics-service/internal/server/data/storage"
)

type counterRepository struct {
	storage storage.Storage
}

func NewCounterRepository(storage storage.Storage) Repository[int64] {
	return &counterRepository{
		storage: storage,
	}
}

func (cr counterRepository) Set(key string, value int64) error {
	return storage.Set[int64](cr.storage, key, value)
}

func (cr counterRepository) Get(key string) (value int64, err error) {
	return storage.Get[int64](cr.storage, key)
}
