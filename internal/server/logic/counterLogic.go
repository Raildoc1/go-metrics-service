package logic

import (
	"errors"
	"fmt"
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
			return fmt.Errorf("%w: getting '%s' failed", err, key)
		}
	}
	newValue := prevValue + delta
	err = cl.repository.Set(key, newValue)
	if err != nil {
		return fmt.Errorf("%w: setting '%s' failed", err, key)
	}
	return nil
}
