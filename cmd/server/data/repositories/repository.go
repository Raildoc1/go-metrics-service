package repositories

type Repository[T any] interface {
	Set(key string, value T) error
	Get(key string) (value T, err error)
}
