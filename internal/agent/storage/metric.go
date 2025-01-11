package storage

type metricValue interface {
	~int64 | ~float64
}

type metric[T metricValue] struct {
	lastCommitedValue *T
	value             T
}

type counter struct {
	metric[int64]
}

type gauge struct {
	metric[float64]
}

func (m *metric[T]) changed() bool {
	return m.lastCommitedValue == nil || (*m.lastCommitedValue) != m.value
}

func (m *metric[T]) commit() {
	tmp := m.value
	m.lastCommitedValue = &tmp
}

func (c *counter) GetUncommitedDelta() (int64, bool) {
	if !c.changed() {
		return 0, false
	}
	if c.lastCommitedValue == nil {
		return c.value, true
	}
	return c.value - *c.lastCommitedValue, true
}

func (g *gauge) GetUncommitedValue() (float64, bool) {
	if !g.changed() {
		return 0, false
	}
	return g.value, true
}
