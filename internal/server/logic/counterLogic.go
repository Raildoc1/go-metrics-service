package logic

import (
	"errors"
	"go-metrics-service/internal/server/data/repositories"
)

type CounterLogic struct {
	repository repositories.Repository[int64]
}

func NewCounterLogic(repository repositories.Repository[int64]) *CounterLogic {
	return &CounterLogic{
		repository: repository,
	}
}

func (cl *CounterLogic) Change(key string, delta int64) error {
	prevValue, err := cl.repository.Get(key)

	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrNotFound):
			prevValue = int64(0)
		default:
			return err
		}
	}

	newValue := prevValue + delta
	return cl.repository.Set(key, newValue)
}
