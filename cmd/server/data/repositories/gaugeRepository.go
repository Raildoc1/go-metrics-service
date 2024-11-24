package repositories

import (
	"go-metrics-service/cmd/server/data/storage"
)

type GaugeRepository struct {
	storage storage.Storage
}

func NewGaugeRepository(storage storage.Storage) GaugeRepository {
	return GaugeRepository{
		storage: storage,
	}
}

func (gr GaugeRepository) Set(key string, value float64) error {
	return storage.Set[float64](gr.storage, key, value)
}

func (gr GaugeRepository) Get(key string) (value float64, err error) {
	return storage.Get[float64](gr.storage, key)
}
