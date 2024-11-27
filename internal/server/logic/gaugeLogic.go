package logic

import (
	"go-metrics-service/internal/server/data/repositories"
)

type GaugeLogic struct {
	repository repositories.Repository[float64]
}

func NewGaugeLogic(repository repositories.Repository[float64]) *GaugeLogic {
	return &GaugeLogic{repository}
}

func (gl *GaugeLogic) Set(key string, value float64) error {
	return gl.repository.Set(key, value)
}
