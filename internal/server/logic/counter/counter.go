package counter

import (
	"fmt"
)

type repository interface {
	Has(key string) bool
	Get(key string) (value int64, err error)
	Set(key string, value int64) error
}

type Counter struct {
	repository repository
}

func New(
	repository repository,
) *Counter {
	return &Counter{
		repository: repository,
	}
}

func (c *Counter) Change(key string, delta int64) error {
	prevValue, err := c.get(key)
	if err != nil {
		return fmt.Errorf("%w: getting counter '%s' failed", err, key)
	}
	newValue := prevValue + delta
	err = c.repository.Set(key, newValue)
	if err != nil {
		return fmt.Errorf("%w: setting counter '%s' failed", err, key)
	}
	return nil
}

func (c *Counter) get(key string) (value int64, err error) {
	if !c.repository.Has(key) {
		return int64(0), nil
	}
	return c.repository.Get(key)
}
