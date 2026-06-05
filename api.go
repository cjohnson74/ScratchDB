package scratchdb

import (
	"os"
	"path/filepath"

	uuid "github.com/google/uuid"
)

type ScratchDB struct {
	dir					string
	keyDir				map[string]keyDirEntry
	activeFile			string
	activeFileHandle	*os.File
	options				Options
	lockFile			*os.File
}

type entry struct {
	crc			uint32
	timeStamp	uint32
	keySize		uint32
	valueSize	uint32
	key			[]byte
	value		[]byte
}

type keyDirEntry struct {
	fileId		string
	valueSize	uint32	
	valuePos	int64
	timeStamp	uint32
}

type Options struct {
	ReadWrite	bool
	SyncOnPut	bool
}

func getActiveFile(directoryName string, fileErr error) (string, error){
	var gerr error
	var activeFile string

	if os.IsExist(fileErr) {
		matches, gerr := filepath.Glob(filepath.Join(directoryName, "active_*"))
		if gerr != nil {
			return "", gerr
		}
		if len(matches) > 0 {
			activeFile = matches[0]
		}
	}

	return activeFile, gerr
}

// TODO 1: Implement only one process can have ReadWrite: true
// You will want to hold the lockFile handle open for the entire
// duration the db is open, this is how you enforce single-writer.
// On Unix systems, you can use locker (flock) on the open handle,
// to prevent another process from aquiring it.
func Open(directoryName string, options Options) (*ScratchDB, error) {
	var err error
	var dataFile string
	var fileHandle *os.File

	err = os.Mkdir(directoryName, 0755)

	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	dataFile, err = getActiveFile(directoryName, err)

	if err != nil {
		return nil, err
	}

	if dataFile == "" {
		dataFile = filepath.Join(directoryName, "active_"+uuid.NewString())
		fileHandle, err = os.OpenFile(dataFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	newDB := ScratchDB{
		dir:				directoryName,
		keyDir:				make(map[string]keyDirEntry),
		activeFile:			dataFile,
		activeFileHandle:	fileHandle,
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
// TODO 2: Close the lockFile
func (db *ScratchDB) Close() error {
	return db.activeFileHandle.Close()
}