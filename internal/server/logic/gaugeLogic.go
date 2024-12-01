package logic

import (
	"fmt"
	"go-metrics-service/internal/server/data/repositories"
)

type GaugeLogic struct {
	repository repositories.Repository[float64]
}

func NewGaugeLogic(repository repositories.Repository[float64]) *GaugeLogic {
	return &GaugeLogic{repository}
}

func (gl *GaugeLogic) Set(key string, value float64) error {
	err := gl.repository.Set(key, value)
	if err != nil {
		return fmt.Errorf("%w: set gauge", err)
	}
	return nil
}
