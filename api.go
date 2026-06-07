package scratchdb

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/flock" // TODO 1: Switch to syscall package
	"github.com/google/uuid"
)

const TOMBSTONE = "-tombstone-"

type ScratchDB struct {
	dir					string
	keyDir				map[string]KeyDirEntry
	activeFile			string
	activeFileHandle	*os.File
	options				Options
	isClosed			bool
	lockFile			*flock.Flock
}

type Entry struct {
	crc			uint32
	timeStamp	uint32
	keySize		uint32
	valueSize	uint32
	key			[]byte
	value		[]byte
}

type KeyDirEntry struct {
	fileId		string
	valueSize	uint32	
	valuePos	uint64
	timeStamp	uint32
}

type Options struct {
	ReadWrite	bool
	SyncOnPut	bool
}

func getFile(directoryName string, filePattern string) (string, error){
	var file string

	matches, err := filepath.Glob(filepath.Join(directoryName, filePattern))
	if err != nil {
		return "", err
	}
	if len(matches) > 0 {
		file = matches[0]
	}

	return file, err
}

func constructEntry(key []byte, value []byte) (valueSize uint32, valuePos uint32, timestamp uint32, entry []byte) {
	timestampBuff := make([]byte, 4)
	keySizeBuff := make([]byte, 2)
	valueSizeBuff := make([]byte, 4)
	crcBuff := make([]byte, 4)
	keySize := uint16(len(key))
	valueSize = uint32(len(value))
	timestamp = uint32(time.Now().Unix())

	binary.BigEndian.PutUint32(timestampBuff, timestamp)
	binary.BigEndian.PutUint16(keySizeBuff, keySize)
	binary.BigEndian.PutUint32(valueSizeBuff, valueSize)

	entry = make([]byte, 0, len(timestampBuff)+len(keySizeBuff)+len(valueSizeBuff)+len(key)+len(value))
	entry = append(entry, timestampBuff...)
	entry = append(entry, keySizeBuff...)
	entry = append(entry, valueSizeBuff...)
	entry = append(entry, key...)
	entry = append(entry, value...)

	checksum := crc32.ChecksumIEEE(entry)
	binary.BigEndian.PutUint32(crcBuff, checksum)

	entry = append(crcBuff, entry...)
	valuePos = uint32(len(entry) - len(value))

	return valueSize, valuePos, timestamp, entry
}

func getActiveFileLen(activeFileHandle *os.File) (fileLen uint32, err error) {
	fi, err := activeFileHandle.Stat()
	if err != nil {
		return 0, err
	}

	return uint32(fi.Size()), nil
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
		keyDir:				make(map[string]KeyDirEntry),
		activeFile:			dataFile,
		activeFileHandle:	activeFileHandle,
		options:			options,
		isClosed:			false,
		lockFile:			lockFile,
	}

	return &newDB, err
}

func (db *ScratchDB) Put(key []byte, value []byte) error {
	if !db.options.ReadWrite {
		return fmt.Errorf("DB does not have write access.")
	} else if db.isClosed {
		return fmt.Errorf("Cannot Put when DB is closed.")
	}

	activeFileLen, err := getActiveFileLen(db.activeFileHandle)
	if err != nil {
		return err
	}

	valueSize, valuePos, timestamp, entry := constructEntry(key, value)
	_, err = db.activeFileHandle.Write(entry)
	if err != nil {
		return err
	}

	if db.options.SyncOnPut {
		db.activeFileHandle.Sync()
	}

	fileId := strings.Replace(db.activeFile, string(db.dir+"/"+"active_"), "", 1)

	db.keyDir[string(key)] = KeyDirEntry{fileId, valueSize, uint64(activeFileLen)+uint64(valuePos), timestamp}
	
	return err
}

func (db *ScratchDB) Get(key []byte) ([]byte, error) {
	keyDirEntry, ok := db.keyDir[string(key)]
	if !ok {
		return nil, fmt.Errorf("failed to get value from keyDir at key: %s", string(key))
	}

	valueBuff := make([]byte, keyDirEntry.valueSize)
	_, err := db.activeFileHandle.Seek(int64(keyDirEntry.valuePos), io.SeekStart)
	if err != nil {
		return nil, err
	}

	_, err = db.activeFileHandle.Read(valueBuff)
	if err != nil {
		return nil, err
	}

	return valueBuff, nil
}

func (db *ScratchDB) Delete(key []byte) error {
	_, ok := db.keyDir[string(key)]
	if !ok {
		return fmt.Errorf("Key does not exist in DB")
	}

	_, _, _, entry := constructEntry(key, []byte(TOMBSTONE))
	_, err := db.activeFileHandle.Write(entry)
	if err != nil {
		return err
	}
	
	delete(db.keyDir, string(key))
	return nil
}

// func (db *ScratchDB) ListKeys() ([][]byte, error)
// func (db *ScratchDB) Sync() error
// func (db *ScratchDB) Fold(fn func(key []byte, value []byte, acc any) any, acc any) (any, error)
// func (db *ScratchDB) Merge(directoryName string) error

func (db *ScratchDB) Close() error {
	if db.lockFile != nil {
		db.lockFile.Unlock()
	}
	db.isClosed = true
	return db.activeFileHandle.Close()
}