package logic

type CounterRepository interface {
	Has(key string) bool
	GetInt64(key string) (value int64, err error)
	SetInt64(key string, value int64) error
}

type GaugeRepository interface {
	SetFloat64(key string, value float64) error
}
