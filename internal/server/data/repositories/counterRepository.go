package repositories

type CounterRepository struct {
	storage Storage
}

func NewCounterRepository(storage Storage) *CounterRepository {
	return &CounterRepository{
		storage: storage,
	}
}

func (cr CounterRepository) Set(key string, value int64) error {
	return set[int64](cr.storage, key, value)
}

func (cr CounterRepository) Get(key string) (value int64, err error) {
	return get[int64](cr.storage, key)
}
