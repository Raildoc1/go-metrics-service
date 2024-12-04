package memory

import "fmt"

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
	fmt.Println(key, ": ", value)
}

func (m *MemStorage) Has(key string) bool {
	_, ok := m.data[key]
	return ok
}

func (m *MemStorage) Get(key string) (any, bool) {
	v, ok := m.data[key]
	return v, ok
}

func (m *MemStorage) GetAll() map[string]any {
	return m.data
}
