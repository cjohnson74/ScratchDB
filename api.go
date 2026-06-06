package scratchdb

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gofrs/flock" // TODO 1: Switch the syscall package
	"github.com/google/uuid"
)

type ScratchDB struct {
	dir					string
	keyDir				map[string]keyDirEntry
	activeFile			string
	activeFileHandle	*os.File
	options				Options
	lockFile			*flock.Flock
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

func getFile(directoryName string, filePattern string) (string, error){
	var gerr error
	var file string

	matches, gerr := filepath.Glob(filepath.Join(directoryName, filePattern))
	if gerr != nil {
		return "", gerr
	}
	if len(matches) > 0 {
		file = matches[0]
	}

	return file, gerr
}

func Open(directoryName string, options Options) (*ScratchDB, error) {
	var err error
	var dataFile string
	var activeFileHandle *os.File
	var lockFile *flock.Flock
	var lock bool

	err = os.Mkdir(directoryName, 0755)

	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	if os.IsExist(err) {
		dataFile, err = getFile(directoryName, "active_*")
		if err != nil {
			return nil, err
		}	
	}

	if dataFile == "" {
		dataFile = filepath.Join(directoryName, "active_"+uuid.NewString())
	}

	if options.ReadWrite {
		lockFilePath := filepath.Join(directoryName, directoryName+".lock")
		lockFile = flock.New(lockFilePath)
		lock, err = lockFile.TryLock()
		if err != nil {
			return nil, err
		}
		if !lock {
			log.Println("Only one DB writer instance can exist")
			return nil, fmt.Errorf("database already open for writing")
		}
	}

	activeFileHandle, err = os.OpenFile(dataFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	newDB := ScratchDB{
		dir:				directoryName,
		keyDir:				make(map[string]keyDirEntry),
		activeFile:			dataFile,
		activeFileHandle:	activeFileHandle,
		options:			options,
		lockFile:			lockFile,
	}

	return &newDB, err
}

// TODO 2: Implement put method
func (db *ScratchDB) Put(key []byte, value []byte) error
func (db *ScratchDB) Get(key []byte) ([]byte, error) 
func (db *ScratchDB) Delete(key []byte) error
func (db *ScratchDB) ListKeys() ([][]byte, error)
func (db *ScratchDB) Fold(fn func(key []byte, value []byte, acc any) any, acc any) (any, error)
func (db *ScratchDB) Merge(directoryName string) error
func (db *ScratchDB) Sync() error
func (db *ScratchDB) Close() error {
	db.lockFile.Unlock()
	return db.activeFileHandle.Close()
}