package storage

type MemStorage struct {
	data map[string]any
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: make(map[string]any),
	}
}

func (ms *MemStorage) Set(key string, value any) {
	ms.data[key] = value
}

func (ms *MemStorage) GetFloat(key string) (float64, bool) {
	return get[float64](key, ms)
}

func (ms *MemStorage) GetInt(key string) (int64, bool) {
	return get[int64](key, ms)
}

func get[T any](key string, ms *MemStorage) (T, bool) {
	value, ok := ms.data[key]
	if !ok {
		var zero T
		return zero, ok
	}
	castedValue, ok := value.(T)
	return castedValue, ok
}
