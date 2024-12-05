package counter

import (
	"fmt"
)

type repository interface {
	Has(key string) bool
	GetInt64(key string) (value int64, err error)
	SetInt64(key string, value int64) error
}

type Counter struct {
	repository repository
}

func New(repository repository) *Counter {
	return &Counter{
		repository: repository,
	}
}

func (c *Counter) Change(key string, delta int64) error {
	var prevValue int64
	if !c.repository.Has(key) {
		prevValue = int64(0)
	} else {
		var err error
		prevValue, err = c.repository.GetInt64(key)
		if err != nil {
			return fmt.Errorf("%w: getting counter '%s' failed", err, key)
		}
	}
	newValue := prevValue + delta
	err := c.repository.SetInt64(key, newValue)
	if err != nil {
		return fmt.Errorf("%w: setting counter '%s' failed", err, key)
	}
	return nil
}
