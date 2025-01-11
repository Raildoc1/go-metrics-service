package storage

type MetricsDiff struct {
	GaugeValues   map[string]float64
	CounterDeltas map[string]int64
}

type Storage struct {
	gauges   map[string]*gauge
	counters map[string]*counter
}

func New() *Storage {
	return &Storage{
		gauges:   make(map[string]*gauge),
		counters: make(map[string]*counter),
	}
}

func (s *Storage) SetCounter(key string, value int64) {
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

	(*val).value = value
}

func (s *Storage) SetGauge(key string, value float64) {
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

	(*val).value = value
}

func (s *Storage) SetGauges(vals map[string]float64) {
	for k, v := range vals {
		s.SetGauge(k, v)
	}
}

func (s *Storage) GetCounter(key string) (int64, bool) {
	if val, ok := s.counters[key]; ok {
		return (*val).value, true
	}
	return 0, false
}

func (s *Storage) GetGauge(key string) (float64, bool) {
	if val, ok := s.gauges[key]; ok {
		return (*val).value, true
	}
	return 0, false
}

func (s *Storage) GetUncommitedData() MetricsDiff {
	res := MetricsDiff{
		GaugeValues:   make(map[string]float64),
		CounterDeltas: make(map[string]int64),
	}
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
	for _, v := range s.counters {
		v.commit()
	}
}
