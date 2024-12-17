package logic

import (
	"fmt"
)

type Gauge struct {
	repository GaugeRepository
	logger     Logger
}

func New(repository GaugeRepository, logger Logger) *Gauge {
	return &Gauge{
		repository: repository,
		logger:     logger,
	}
}

func (g *Gauge) Set(key string, value float64) error {
	g.logger.Debugln("setting gauge ", key, " ", value)
	err := g.repository.SetFloat64(key, value)
	if err != nil {
		return fmt.Errorf("%w: setting gauge '%s' failed", err, key)
	}
	return nil
}
