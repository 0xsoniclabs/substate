package db

import (
	"io"

	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/syndtr/goleveldb/leveldb"
	ldbiterator "github.com/syndtr/goleveldb/leveldb/iterator"
)

//go:generate mockgen -source=base_db.go -destination=./base_db_mock.go -package=db

// DbAdapter defines the interface for a database adapter that provides
// basic database operations such as Put, Get, Delete, and iteration.
type DbAdapter interface {
	// Delete removes the key-value pair associated with the given key.
	// wo specifies the write options.
	Delete(key []byte, wo *opt.WriteOptions) error

	// Put inserts the given key-value pair into the database.
	// wo specifies the write options.
	Put(key []byte, value []byte, wo *opt.WriteOptions) error

	// Close closes the database, releasing any resources held.
	Close() error

	// Has checks if the database contains the given key.
	// ro specifies the read options.
	Has(key []byte, ro *opt.ReadOptions) (bool, error)

	// CompactRange compacts the database over the given key range.
	CompactRange(u util.Range) error

	// Get retrieves the value associated with the given key.
	// ro specifies the read options.
	Get(key []byte, ro *opt.ReadOptions) ([]byte, error)

	// GetProperty retrieves the value of a database property.
	GetProperty(property string) (string, error)

	// NewIterator creates a new iterator over the given key range.
	// ro specifies the read options.
	NewIterator(r *util.Range, ro *opt.ReadOptions) ldbiterator.Iterator

	// Write writes a batch of operations to the database.
	// wo specifies the write options.
	Write(batch *leveldb.Batch, wo *opt.WriteOptions) error

	// Stats retrieves the database statistics.
	Stats(s *leveldb.DBStats) error
}

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

	// GetSubstateEncoding returns the encoding schema in use.
	GetSubstateEncoding() SubstateEncodingSchema

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

	// GetBackend returns the database backend.
	GetBackend() DbAdapter

	hasKeyValuesFor(prefix []byte, start []byte) bool

	binarySearchForLastPrefixKey(lastKeyPrefix []byte) (byte, error)

	newIterator(r *util.Range) ldbiterator.Iterator

	stats(stats *leveldb.DBStats) error
}
