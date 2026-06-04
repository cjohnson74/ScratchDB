package scratchdb

import (
	"os"
	"path/filepath"
	uuid "github.com/google/uuid"
)

type ScratchDB struct {
	dir				string
	keyDir			map[string][]byte
	activeFile		string
	options			Options
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

func Open(directoryName string, options Options) (*ScratchDB, error) {
	var err error

	err = os.Mkdir(directoryName, 0755)

	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	dataFile := filepath.Join(directoryName, uuid.NewString())

	err = createEmptyFile(dataFile)

	if err != nil {
		return nil, err
	}

	newDB := ScratchDB{
		dir:				directoryName,
		keyDir:				make(map[string][]byte),
		activeFile:			dataFile,
		options:			options,
	}

	return &newDB, err
}

func (db *ScratchDB) Put(key []byte, value []byte) error
func (db *ScratchDB) Get(key []byte) ([]byte, error) 
func (db *ScratchDB) Delete(key []byte) error
func (db *ScratchDB) ListKeys() ([][]byte, error)
func (db *ScratchDB) Fold(fn func(key []byte, value []byte, acc any) any, acc any) (any, error)
func (db *ScratchDB) Merge(directoryName string) error
func (db *ScratchDB) Sync() error
func (db *ScratchDB) Close() error