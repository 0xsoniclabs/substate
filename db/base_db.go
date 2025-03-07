package db

import (
	"fmt"
	"io"
	"os"

	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/syndtr/goleveldb/leveldb"
	ldbiterator "github.com/syndtr/goleveldb/leveldb/iterator"
)

// KeyValueWriter wraps the Put method of a backing data store.
type KeyValueWriter interface {
	// Put inserts the given value into the key-value data store.
	Put(key []byte, value []byte) error

	// Delete removes the key from the key-value data store.
	Delete(key []byte) error
}

type BaseDB interface {
	KeyValueWriter

	io.Closer

	// Has returns true if the baseDB does contain the given key.
	Has([]byte) (bool, error)

	// Get gets the value for the given key.
	Get([]byte) ([]byte, error)

	// NewBatch creates a write-only database that buffers changes to its host db
	// until a final write is called.
	NewBatch() Batch

	// NewIterator creates a binary-alphabetical iterates over a subset
	// of database content with a particular key prefix, starting at a particular
	// initial key (or after, if it does not exist).
	//
	// Note: This method assumes that the prefix is NOT part of the start, so there's
	// no need for the caller to prepend the prefix to the start
	NewIterator(prefix []byte, start []byte) ldbiterator.Iterator

	// Stat returns a particular internal stat of the database.
	Stat(property string) (string, error)

	// Compact flattens the underlying data store for the given key range. In essence,
	// deleted and overwritten versions are discarded, and the data is rearranged to
	// reduce the cost of operations needed to access them.
	//
	// A nil start is treated as a key before all keys in the data store; a nil limit
	// is treated as a key after all keys in the data store. If both is nil then it
	// will compact entire data store.
	Compact(start []byte, limit []byte) error

	// Close closes the DB. This will also release any outstanding snapshot,
	// abort any in-flight compaction and discard open transaction.
	//
	// Note:
	// It is not safe to close a DB until all outstanding iterators are released.
	// It is valid to call Close multiple times.
	// Other methods should not be called after the DB has been closed.
	Close() error

	// getBackend returns the database backend.
	getBackend() *leveldb.DB
}

// NewDefaultBaseDB creates new instance of BaseDB with default options.
func NewDefaultBaseDB(path string) (BaseDB, error) {
	return newBaseDB(path, nil, nil, nil)
}

// NewBaseDB creates new instance of BaseDB with customizable options.
// Note: Any of three options is nillable. If that's the case a default value for the option is set.
func NewBaseDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (BaseDB, error) {
	return newBaseDB(path, o, wo, ro)
}

func MakeDefaultBaseDBFromBaseDB(db BaseDB) BaseDB {
	return &baseDB{backend: db.getBackend()}
}

// NewReadOnlyBaseDB creates a new instance of read-only BaseDB.
func NewReadOnlyBaseDB(path string) (BaseDB, error) {
	return newBaseDB(path, &opt.Options{ReadOnly: true}, nil, nil)
}

// OpenBaseDB opens existing database. If it does not exists error is returned instead.
func OpenBaseDB(path string) (BaseDB, error) {
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return NewDefaultBaseDB(path)
}

func newBaseDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (*baseDB, error) {
	b, err := leveldb.OpenFile(path, o)
	if err != nil {
		return nil, fmt.Errorf("cannot open leveldb; %w", err)
	}
	return &baseDB{
		backend: b,
		wo:      wo,
		ro:      ro,
	}, nil
}

// baseDB implements method needed by all three types of DBs.
type baseDB struct {
	backend *leveldb.DB
	wo      *opt.WriteOptions
	ro      *opt.ReadOptions
}

func (db *baseDB) getBackend() *leveldb.DB {
	return db.backend
}

func (db *baseDB) Put(key []byte, value []byte) error {
	return db.backend.Put(key, value, db.wo)
}

func (db *baseDB) Delete(key []byte) error {
	return db.backend.Delete(key, db.wo)
}

func (db *baseDB) Close() error {
	return db.backend.Close()
}

func (db *baseDB) Has(key []byte) (bool, error) {
	return db.backend.Has(key, db.ro)
}

func (db *baseDB) Get(key []byte) ([]byte, error) {
	return db.backend.Get(key, db.ro)
}

func (db *baseDB) NewBatch() Batch {
	return newBatch(db.backend)
}

// newIterator returns iterator which iterates over values depending on the prefix.
// Note: If prefix is nil, everything is iterated.
func (db *baseDB) NewIterator(prefix []byte, start []byte) ldbiterator.Iterator {
	r := util.BytesPrefix(prefix)
	r.Start = append(r.Start, start...)
	return db.backend.NewIterator(r, db.ro)
}

func (db *baseDB) Stat(property string) (string, error) {
	return db.backend.GetProperty(property)
}

func (db *baseDB) Compact(start []byte, limit []byte) error {
	return db.backend.CompactRange(util.Range{Start: start, Limit: limit})
}

func (db *baseDB) hasKeyValuesFor(prefix []byte, start []byte) bool {
	iter := db.NewIterator(prefix, start)
	defer iter.Release()
	return iter.Next()
}

func (db *baseDB) binarySearchForLastPrefixKey(lastKeyPrefix []byte) (byte, error) {
	var min uint16 = 0
	var max uint16 = 255

	startIndex := make([]byte, 1)

	for max-min > 1 {
		searchHalf := (max + min) / 2
		startIndex[0] = byte(searchHalf)
		if db.hasKeyValuesFor(lastKeyPrefix, startIndex) {
			min = searchHalf
		} else {
			max = searchHalf
		}
	}

	// shouldn't occure
	if max-min == 0 {
		return 0, fmt.Errorf("undefined behaviour in GetLastSubstate search; max - min == 0")
	}

	startIndex[0] = byte(min)
	if db.hasKeyValuesFor(lastKeyPrefix, startIndex) {
		startIndex[0] = byte(max)
		if db.hasKeyValuesFor(lastKeyPrefix, startIndex) {
			return byte(max), nil
		} else {
			return byte(min), nil
		}
	} else {
		return 0, fmt.Errorf("undefined behaviour in GetLastSubstate search")
	}
}
