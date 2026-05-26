package database

type Store interface {
	Get(key []byte, value []byte) ([]byte, error)
	Put(key, value []byte) error
	Delete(key []byte) error
}