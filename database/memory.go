package database

import (
	"maps"
)

type MemoryStore struct {
	data map[string][]byte
}