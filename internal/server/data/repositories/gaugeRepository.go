package repositories

import (
	"go-metrics-service/internal/server/data/storage"
)

type gaugeRepository struct {
	storage storage.Storage
}

func NewGaugeRepository(storage storage.Storage) Repository[float64] {
	return &gaugeRepository{
		storage: storage,
	}
}

func (gr gaugeRepository) Set(key string, value float64) error {
	return storage.Set[float64](gr.storage, key, value)
}

func (gr gaugeRepository) Get(key string) (value float64, err error) {
	return storage.Get[float64](gr.storage, key)
}
