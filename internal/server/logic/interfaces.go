package logic

type CounterRepository interface {
	Has(key string) (bool, error)
	SetCounter(key string, value int64) error
	GetCounter(key string) (int64, error)
}

type GaugeRepository interface {
	SetGauge(key string, value float64) error
}
