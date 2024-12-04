package repositories

type GaugeRepository struct {
	storage Storage
}

func NewGaugeRepository(storage Storage) *GaugeRepository {
	return &GaugeRepository{
		storage: storage,
	}
}

func (gr GaugeRepository) Set(key string, value float64) error {
	return set[float64](gr.storage, key, value)
}

func (gr GaugeRepository) Has(key string) bool {
	return gr.Has(key)
}

func (gr GaugeRepository) Get(key string) (value float64, err error) {
	return get[float64](gr.storage, key)
}
