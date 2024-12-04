package gauge

import (
	"fmt"
)

type repository interface {
	Set(key string, value float64) error
}

type Gauge struct {
	repository repository
}

func New(
	repository repository,
) *Gauge {
	return &Gauge{repository}
}

func (gl *Gauge) Set(key string, value float64) error {
	err := gl.repository.Set(key, value)
	if err != nil {
		return fmt.Errorf("%w: setting gauge '%s' failed", err, key)
	}
	return nil
}
