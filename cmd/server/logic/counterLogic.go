package logic

import "go-metrics-service/cmd/server/storage"

type CounterLogic struct {
	storage storage.Storage
}

func NewCounterLogic(storage storage.Storage) *CounterLogic {
	return &CounterLogic{storage}
}

func (cl *CounterLogic) Change(key string, delta int64) {
	extendedKey := storage.CounterKeyPrefix + key
	prevValue, ok := cl.storage.GetInt(extendedKey)

	if !ok {
		prevValue = int64(0)
	}

	newValue := prevValue + delta
	cl.storage.Set(extendedKey, newValue)
}
