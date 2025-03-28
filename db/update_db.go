package db

import (
	"encoding/binary"
	"fmt"

	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	UpdateDBPrefix = "2s" // UpdateDBPrefix + block (64-bit) + tx (64-bit) -> substateRLP
)

// UpdateDB represents a CodeDB with in which the UpdateSet is inserted.
//
//go:generate mockgen -source=update_db.go -destination=./update_db_mock.go -package=db
type UpdateDB interface {
	CodeDB

	// SetSubstateEncoding sets the runtime encoding/decoding
	SetSubstateEncoding(schema SubstateEncodingSchema) error

	// GetSubstateEncoding returns the encoding schema in use.
	GetSubstateEncoding() SubstateEncodingSchema

	// GetFirstKey returns block number of first UpdateSet. It returns an error if no UpdateSet is found.
	GetFirstKey() (uint64, error)

	// GetLastKey returns block number of last UpdateSet. It returns an error if no UpdateSet is found.
	GetLastKey() (uint64, error)

	// HasUpdateSet returns true if there is an UpdateSet on given block.
	HasUpdateSet(block uint64) (bool, error)

	// GetUpdateSet returns UpdateSet for given block. If there is not an UpdateSet for the block, nil is returned.
	GetUpdateSet(block uint64) (*updateset.UpdateSet, error)

	// PutUpdateSet inserts the UpdateSet with deleted accounts into the DB assigned to given block.
	PutUpdateSet(updateSet *updateset.UpdateSet, deletedAccounts []types.Address) error

	// DeleteUpdateSet deletes UpdateSet for given block. It returns an error if there is no UpdateSet on given block.
	DeleteUpdateSet(block uint64) error

	NewUpdateSetIterator(start, end uint64) IIterator[*updateset.UpdateSet]

	PutMetadata(interval, size uint64) error
}

// NewDefaultUpdateDB creates new instance of UpdateDB with default options.
func NewDefaultUpdateDB(path string) (UpdateDB, error) {
	return newUpdateDB(path, nil, nil, nil)
}

// NewUpdateDB creates new instance of UpdateDB with customizable options.
// Note: Any of three options is nillable. If that's the case a default value for the option is set.
func NewUpdateDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (UpdateDB, error) {
	return newUpdateDB(path, o, wo, ro)
}

func MakeDefaultUpdateDBFromBaseDB(db BaseDB) UpdateDB {
	encoding, err := newUpdateSetEncoding(DefaultEncodingSchema)
	if err != nil {
		// This should not happen
		panic(fmt.Sprintf("failed to create default update-db encoding: %v", err))
	}
	return &updateDB{
		&codeDB{&baseDB{backend: db.GetBackend()}},
		*encoding,
	}
}

// NewReadOnlyUpdateDB creates a new instance of read-only UpdateDB.
func NewReadOnlyUpdateDB(path string) (UpdateDB, error) {
	return newUpdateDB(path, &opt.Options{ReadOnly: true}, nil, nil)
}

func newUpdateDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (*updateDB, error) {
	base, err := newCodeDB(path, o, wo, ro)
	if err != nil {
		return nil, err
	}
	encoding, err := newUpdateSetEncoding(DefaultEncodingSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to create default update-db encoding: %v", err)
	}
	return &updateDB{
		base,
		*encoding,
	}, nil
}

type updateDB struct {
	CodeDB
	encoding updateSetEncoding
}

func (db *updateDB) GetFirstKey() (uint64, error) {
	r := util.BytesPrefix([]byte(UpdateDBPrefix))

	iter := db.newIterator(r)
	defer iter.Release()

	exist := iter.First()
	if exist {
		firstBlock, err := DecodeUpdateSetKey(iter.Key())
		if err != nil {
			return 0, fmt.Errorf("cannot decode updateset key; %v", err)
		}
		return firstBlock, nil
	}
	return 0, leveldb.ErrNotFound
}

func (db *updateDB) GetLastKey() (uint64, error) {
	r := util.BytesPrefix([]byte(UpdateDBPrefix))

	iter := db.newIterator(r)
	defer iter.Release()

	exist := iter.Last()
	if exist {
		lastBlock, err := DecodeUpdateSetKey(iter.Key())
		if err != nil {
			return 0, fmt.Errorf("cannot decode updateset key; %v", err)
		}
		return lastBlock, nil
	}
	return 0, leveldb.ErrNotFound
}

func (db *updateDB) HasUpdateSet(block uint64) (bool, error) {
	key := UpdateDBKey(block)
	return db.Has(key)
}

func (db *updateDB) GetUpdateSet(block uint64) (*updateset.UpdateSet, error) {
	key := UpdateDBKey(block)
	value, err := db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("cannot get updateset block: %v, key %v; %w", block, key, err)
	}

	if value == nil {
		return nil, nil
	}

	// decode value
	data, err := db.encoding.decode(block, db.GetCode, value)
	if err != nil {
		return nil, fmt.Errorf("cannot decode update-set block: %v, key %v; %w", block, key, err)
	}
	return data, nil
}

func (db *updateDB) PutUpdateSet(updateSet *updateset.UpdateSet, deletedAccounts []types.Address) error {
	// put deployed/creation code
	for _, account := range updateSet.WorldState {
		err := db.PutCode(account.Code)
		if err != nil {
			return err
		}
	}

	key := UpdateDBKey(updateSet.Block)
	value, err := db.encoding.encode(*updateSet, deletedAccounts)
	if err != nil {
		return fmt.Errorf("cannot encode update-set; %v", err)
	}

	return db.Put(key, value)
}

func (db *updateDB) DeleteUpdateSet(block uint64) error {
	key := UpdateDBKey(block)
	return db.Delete(key)
}

func (db *updateDB) NewUpdateSetIterator(start, end uint64) IIterator[*updateset.UpdateSet] {
	iter := newUpdateSetIterator(db, start, end, db.encoding.decode)

	iter.start(0)

	return iter
}

func DecodeUpdateSetKey(key []byte) (block uint64, err error) {
	prefix := UpdateDBPrefix
	if len(key) != len(prefix)+8 {
		err = fmt.Errorf("invalid length of updateset key: %v", len(key))
		return
	}
	if p := string(key[:len(prefix)]); p != prefix {
		err = fmt.Errorf("invalid prefix of updateset key: %#x", p)
		return
	}
	blockTx := key[len(prefix):]
	block = binary.BigEndian.Uint64(blockTx[0:8])
	return
}

func UpdateDBKey(block uint64) []byte {
	prefix := []byte(UpdateDBPrefix)
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx[0:8], block)
	return append(prefix, blockTx...)
}
