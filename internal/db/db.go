package db

import (
	"log"

	"github.com/tharunn0/blitzdb/internal/db/storage"
)

type DB struct {
	storage *storage.Storage
}

func InitDB() *DB {
	return &DB{
		storage: storage.Init(),
	}
}

func (db *DB) Set(key string, value []byte, ttl uint64) {
	db.storage.Write(key, value)
}

func (db *DB) Get(key string) ([]byte, bool) {
	val, err := db.storage.Read(key)
	if err != nil {
		log.Println("error :", err)
		return nil, false
	}
	return val, true
}

func (db *DB) Delete(key string) {
	db.storage.Delete(key)
}

func (db *DB) Janitor() { log.Println("todo Janitor") }
