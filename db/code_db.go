package db

import (
	"errors"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	ldbiterator "github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/types/hash"
)

const CodeDBPrefix = "1c" // CodeDBPrefix + codeHash (256-bit) -> code

// CodeDB is a wrapper around BaseDB. It extends it with Has/Get/PutCode functions.
//
//go:generate mockgen -source=code_db.go -destination=./code_db_mock.go -package=db
type CodeDB interface {
	BaseDB

	// HasCode returns true if the DB does contain given code hash.
	HasCode(types.Hash) (bool, error)

	// GetCode gets the code for the given hash.
	GetCode(types.Hash) ([]byte, error)

	// PutCode creates hash for given code and inserts it into the DB.
	PutCode([]byte) error

	// DeleteCode deletes the code for given hash.
	DeleteCode(types.Hash) error
}

// NewDefaultCodeDB creates new instance of CodeDB with default options.
func NewDefaultCodeDB(path string) (CodeDB, error) {
	return newCodeDB(path, nil, nil, nil)
}

// NewCodeDB creates new instance of CodeDB with customizable options.
// Note: Any of three options is nillable. If that's the case a default value for the option is set.
func NewCodeDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (CodeDB, error) {
	return newCodeDB(path, o, wo, ro)
}

func MakeDefaultCodeDBFromBaseDB(db BaseDB) CodeDB {
	return &codeDB{db.GetBackend(), nil, nil}
}

// NewReadOnlyCodeDB creates a new instance of read-only CodeDB.
func NewReadOnlyCodeDB(path string) (CodeDB, error) {
	return newCodeDB(path, &opt.Options{ReadOnly: true}, nil, nil)
}

func newCodeDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (*codeDB, error) {
	b, err := leveldb.OpenFile(path, o)
	if err != nil {
		return nil, fmt.Errorf("cannot open leveldb; %w", err)
	}
	return &codeDB{
		backend: b,
		wo:      wo,
		ro:      ro,
	}, nil
}

type codeDB struct {
	backend DbAdapter
	wo      *opt.WriteOptions
	ro      *opt.ReadOptions
}

var ErrorEmptyHash = errors.New("give hash is empty")

// HasCode returns true if the baseDB does contain given code hash.
func (db *codeDB) HasCode(codeHash types.Hash) (bool, error) {
	if codeHash.IsEmpty() {
		return false, ErrorEmptyHash
	}

	key := CodeDBKey(codeHash)
	has, err := db.Has(key)
	if err != nil {
		return false, err
	}
	return has, nil
}

// GetCode gets the code for the given hash.
func (db *codeDB) GetCode(codeHash types.Hash) ([]byte, error) {
	if codeHash.IsEmpty() {
		return nil, ErrorEmptyHash
	}

	key := CodeDBKey(codeHash)
	code, err := db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("cannot get code %s: %w", codeHash, err)
	}
	return code, nil
}

// PutCode creates hash for given code and inserts it into the baseDB.
func (db *codeDB) PutCode(code []byte) error {
	codeHash := hash.Keccak256Hash(code)
	key := CodeDBKey(codeHash)
	err := db.Put(key, code)
	if err != nil {
		return fmt.Errorf("cannot put code %s: %w", codeHash, err)
	}

	return nil
}

// DeleteCode deletes the code for the given hash.
func (db *codeDB) DeleteCode(codeHash types.Hash) error {
	if codeHash.IsEmpty() {
		return ErrorEmptyHash
	}

	key := CodeDBKey(codeHash)
	err := db.Delete(key)
	if err != nil {
		return fmt.Errorf("cannot delete code %s: %w", codeHash, err)
	}
	return nil
}

func (db *codeDB) stats(stats *leveldb.DBStats) error {
	return db.backend.Stats(stats)
}

func (db *codeDB) GetBackend() DbAdapter {
	return db.backend
}

func (db *codeDB) Put(key []byte, value []byte) error {
	return db.backend.Put(key, value, db.wo)
}

func (db *codeDB) Delete(key []byte) error {
	return db.backend.Delete(key, db.wo)
}

func (db *codeDB) Close() error {
	return db.backend.Close()
}

func (db *codeDB) Has(key []byte) (bool, error) {
	return db.backend.Has(key, db.ro)
}

func (db *codeDB) Get(key []byte) ([]byte, error) {
	return db.backend.Get(key, db.ro)
}

func (db *codeDB) NewBatch() Batch {
	return newBatch(db.backend)
}

func (db *codeDB) NewIterator(prefix []byte, start []byte) ldbiterator.Iterator {
	r := util.BytesPrefix(prefix)
	r.Start = append(r.Start, start...)
	return db.backend.NewIterator(r, db.ro)
}

func (db *codeDB) newIterator(r *util.Range) ldbiterator.Iterator {
	return db.backend.NewIterator(r, db.ro)
}

func (db *codeDB) Stat(property string) (string, error) {
	return db.backend.GetProperty(property)
}

func (db *codeDB) Compact(start []byte, limit []byte) error {
	return db.backend.CompactRange(util.Range{Start: start, Limit: limit})
}

func (db *codeDB) hasKeyValuesFor(prefix []byte, start []byte) bool {
	iter := db.NewIterator(prefix, start)
	defer iter.Release()
	return iter.Next()
}

func (db *codeDB) binarySearchForLastPrefixKey(lastKeyPrefix []byte) (byte, error) {
	var mMin uint16 = 0
	var mMax uint16 = 255

	startIndex := make([]byte, 1)

	for mMax-mMin > 1 {
		searchHalf := (mMax + mMin) / 2
		startIndex[0] = byte(searchHalf)
		if db.hasKeyValuesFor(lastKeyPrefix, startIndex) {
			mMin = searchHalf
		} else {
			mMax = searchHalf
		}
	}

	// shouldn't occur
	if mMax-mMin == 0 {
		return 0, fmt.Errorf("undefined behaviour in GetLastSubstate search; mMax - mMin == 0")
	}

	startIndex[0] = byte(mMin)
	if db.hasKeyValuesFor(lastKeyPrefix, startIndex) {
		startIndex[0] = byte(mMax)
		if db.hasKeyValuesFor(lastKeyPrefix, startIndex) {
			return byte(mMax), nil
		} else {
			return byte(mMin), nil
		}
	} else {
		return 0, fmt.Errorf("undefined behaviour in GetLastSubstate search")
	}
}

// CodeDBKey returns CodeDBPrefix with appended
// codeHash creating key used in baseDB for Codes.
func CodeDBKey(codeHash types.Hash) []byte {
	prefix := []byte(CodeDBPrefix)
	return append(prefix, codeHash[:]...)
}

// DecodeCodeDBKey decodes key created by CodeDBKey back to hash.
func DecodeCodeDBKey(key []byte) (codeHash types.Hash, err error) {
	prefix := CodeDBPrefix
	if len(key) != len(prefix)+32 {
		err = fmt.Errorf("invalid length of code db key: %v", len(key))
		return
	}
	if p := string(key[:2]); p != prefix {
		err = fmt.Errorf("invalid prefix of code db key: %#x", p)
		return
	}
	var h types.Hash
	h.SetBytes(key[len(prefix):])
	codeHash = types.BytesToHash(key[len(prefix):])
	return
}
