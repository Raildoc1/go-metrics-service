package storage

type MemStorage struct {
	data map[string]any
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: make(map[string]any),
	}
}

func (m *MemStorage) Set(key string, value any) {
	m.data[key] = value
}

func (m *MemStorage) Get(key string) (any, bool) {
	v, ok := m.data[key]
	return v, ok
}
