package scratchdb

type ScratchDB struct {}

type Options struct {
	ReadWrite bool
	SyncOnPut bool
}

func Open(directoryName string, options Options) (*ScratchDB, error)
func (db *ScratchDB) Get(key []byte) ([]byte, error)
func (db *ScratchDB) Put(key []byte, value []byte) error
func (db *ScratchDB) Delete(key []byte) error
func (db *ScratchDB) ListKeys() ([][]byte, error)
func (db *ScratchDB) Fold(fn func(key []byte, value []byte, acc any) any, acc any) (any, error)
func (db *ScratchDB) Merge(directoryName string) error
func (db *ScratchDB) Sync() error
func (db *ScratchDB) Close() error
