package storage

import (
	"sync"
)

type MetricsDiff struct {
	GaugeValues   map[string]float64
	CounterDeltas map[string]int64
}

type Storage struct {
	gauges   map[string]*gauge
	gMutex   *sync.RWMutex
	counters map[string]*counter
	cMutex   *sync.RWMutex
}

func New() *Storage {
	return &Storage{
		gauges:   make(map[string]*gauge),
		gMutex:   &sync.RWMutex{},
		counters: make(map[string]*counter),
		cMutex:   &sync.RWMutex{},
	}
}

func (s *Storage) SetCounter(key string, value int64) {
	s.cMutex.Lock()
	defer s.cMutex.Unlock()

	val, ok := s.counters[key]
	if !ok {
		s.counters[key] = &counter{
			metric: metric[int64]{
				lastCommitedValue: nil,
				value:             value,
			},
		}
		return
	}

	val.value = value
}

func (s *Storage) SetGauge(key string, value float64) {
	s.gMutex.Lock()
	defer s.gMutex.Unlock()

	val, ok := s.gauges[key]
	if !ok {
		s.gauges[key] = &gauge{
			metric: metric[float64]{
				lastCommitedValue: nil,
				value:             value,
			},
		}
		return
	}

	val.value = value
}

func (s *Storage) SetGauges(vals map[string]float64) {
	for k, v := range vals {
		s.SetGauge(k, v)
	}
}

func (s *Storage) GetCounter(key string) (int64, bool) {
	s.cMutex.RLock()
	defer s.cMutex.RUnlock()

	if val, ok := s.counters[key]; ok {
		return val.value, true
	}
	return 0, false
}

func (s *Storage) GetGauge(key string) (float64, bool) {
	s.gMutex.RLock()
	defer s.gMutex.RUnlock()

	if val, ok := s.gauges[key]; ok {
		return val.value, true
	}
	return 0, false
}

func (s *Storage) GetUncommitedData() MetricsDiff {
	res := MetricsDiff{
		GaugeValues:   make(map[string]float64),
		CounterDeltas: make(map[string]int64),
	}

	s.gMutex.Lock()
	s.cMutex.Lock()

	for k, v := range s.gauges {
		if val, ok := v.GetUncommitedValue(); ok {
			res.GaugeValues[k] = val
		}
	}
	for k, v := range s.counters {
		if delta, ok := v.GetUncommitedDelta(); ok {
			res.CounterDeltas[k] = delta
		}
	}

	return res
}

func (s *Storage) Commit() {
	for _, v := range s.gauges {
		v.commit()
	}
	s.gMutex.Unlock()
	for _, v := range s.counters {
		v.commit()
	}
	s.cMutex.Unlock()
}

func (s *Storage) ConsumeUncommitedCounters() map[string]int64 {
	s.cMutex.Lock()
	defer s.cMutex.Unlock()

	deltas := make(map[string]int64)

	for k, v := range s.counters {
		if delta, ok := v.GetUncommitedDelta(); ok {
			deltas[k] = delta
		}
		v.commit()
	}

	return deltas
}

func (s *Storage) ConsumeUncommitedGauges() map[string]float64 {
	s.gMutex.Lock()
	defer s.gMutex.Unlock()

	values := make(map[string]float64)

	for k, v := range s.gauges {
		if value, ok := v.GetUncommitedValue(); ok {
			values[k] = value
		}
		v.commit()
	}

	return values
}
