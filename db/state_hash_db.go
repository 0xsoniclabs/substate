package db

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/0xsoniclabs/substate/types"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const StateHashDBPrefix = "dbh" // StateHashDBPrefix + 0x + blockNum(hex) -> StateHash

// StateHashDB is a wrapper around CodeDB. It extends it with Has/Get/Put/DeleteSubstate functions.
//
//go:generate mockgen -source=substate_db.go -destination=./substate_db_mock.go -package=db
type StateHashDB interface {
	BaseDB

	PutStateHash(blockNumber uint64, stateRoot []byte) error
}

// NewDefaultStateHashDB creates new instance of StateHashDB with default options.
func NewDefaultStateHashDB(path string) (StateHashDB, error) {
	return newStateHashDB(path, nil, nil, nil)
}

// NewStateHashDB creates new instance of StateHashDB with customizable options.
// Note: Any of three options is nillable. If that's the case a default value for the option is set.
func NewStateHashDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (StateHashDB, error) {
	return newStateHashDB(path, o, wo, ro)
}

func MakeDefaultStateHashDB(db *leveldb.DB) StateHashDB {
	return &stateHashDB{&baseDB{backend: db}}
}

func MakeDefaultStateHashDBFromBaseDB(db BaseDB) StateHashDB {
	return &stateHashDB{&baseDB{backend: db.getBackend()}}
}

// NewReadOnlyStateHashDB creates a new instance of read-only StateHashDB.
func NewReadOnlyStateHashDB(path string) (StateHashDB, error) {
	return newStateHashDB(path, &opt.Options{ReadOnly: true}, nil, nil)
}

func MakeStateHashDB(db *leveldb.DB, wo *opt.WriteOptions, ro *opt.ReadOptions) StateHashDB {
	return &stateHashDB{&baseDB{backend: db, wo: wo, ro: ro}}
}

func newStateHashDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (*stateHashDB, error) {
	base, err := newBaseDB(path, o, wo, ro)
	if err != nil {
		return nil, err
	}
	return &stateHashDB{base}, nil
}

type stateHashDB struct {
	BaseDB
}

func (db *stateHashDB) PutStateHash(blockNumber uint64, stateRoot []byte) error {
	hex := strconv.FormatUint(blockNumber, 16)
	fullPrefix := StateHashDBPrefix + "0x" + hex
	err := db.Put([]byte(fullPrefix), stateRoot)
	if err != nil {
		return fmt.Errorf("unable to put state hash for block %d: %v", blockNumber, err)
	}
	return nil
}

func (db *stateHashDB) PutStateHashString(blockNumber string, stateRoot string) error {
	fullPrefix := StateHashDBPrefix + blockNumber
	err := db.Put([]byte(fullPrefix), types.Hex2Bytes(strings.TrimPrefix(stateRoot, "0x")))
	if err != nil {
		return fmt.Errorf("unable to put state hash for block %s: %v", blockNumber, err)
	}
	return nil
}

func (db *stateHashDB) GetStateHash(number int) (types.Hash, error) {
	hex := strconv.FormatUint(uint64(number), 16)
	stateRoot, err := db.Get([]byte(StateHashDBPrefix + "0x" + hex))
	if err != nil {
		return types.Hash{}, err
	}

	if stateRoot == nil {
		return types.Hash{}, nil
	}

	return types.Hash(stateRoot), nil
}
