package scratchdb

import (
	"os"
	uuid "github.com/google/uuid"
	"maps"
)

type ScratchDB struct {
	// TODO 1: Fill in ScratchDB struct. What does ScratchDB need to hold internally?
	dir				string
	keyDir			map[uuid.UUID][]byte
	activeFile		string
	options				Options
}

type Options struct {
	ReadWrite	bool
	SyncOnPut	bool
}

func createEmptyFile(name string) error {
	d := []byte("")
	err := os.WriteFile(name, d, 0644)
	return err
}

// TODO 2: implement Open function
// TODO 3: Add unit tests of Open function
func Open(directoryName string, options Options) (*ScratchDB, error) {
	var err error

	err = os.Mkdir(directoryName, 0755)

	if err != nil {
		return nil, err
	}

	dataFile := uuid.NewString()

	err = createEmptyFile(dataFile)

	if err != nil {
		return nil, err
	}

	newDB := ScratchDB{
		dir:				directoryName,
		keyDir:				make(map[uuid.UUID][]byte),
		activeFile:			dataFile,
		options:			options,
	}

	return &newDB, err
}

func (db *ScratchDB) Get(key []byte) ([]byte, error)
func (db *ScratchDB) Put(key []byte, value []byte) error
func (db *ScratchDB) Delete(key []byte) error
func (db *ScratchDB) ListKeys() ([][]byte, error)
func (db *ScratchDB) Fold(fn func(key []byte, value []byte, acc any) any, acc any) (any, error)
func (db *ScratchDB) Merge(directoryName string) error
func (db *ScratchDB) Sync() error
func (db *ScratchDB) Close() error