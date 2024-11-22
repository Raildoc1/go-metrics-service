package storage

type Storage interface {
	Set(key string, value any)
	GetFloat(key string) (value float64, ok bool)
	GetInt(key string) (value int64, ok bool)
}
