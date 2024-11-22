package logic

import "go-metrics-service/cmd/server/storage"

type GaugeLogic struct {
	storage storage.Storage
}

func NewGaugeLogic(storage storage.Storage) *GaugeLogic {
	return &GaugeLogic{storage}
}

func (gl *GaugeLogic) Set(key string, value float64) {
	extendedKey := storage.GaugeKeyPrefix + key
	gl.storage.Set(extendedKey, value)
}
