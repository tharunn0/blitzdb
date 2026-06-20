package store

type Store interface {
	Set(key string, value []byte, ttl uint64)
	Get(key string) ([]byte, bool)
	Delete(key string)
	Janitor()
}
