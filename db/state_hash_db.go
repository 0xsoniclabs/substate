package db

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/0xsoniclabs/substate/types"
	"github.com/status-im/keycard-go/hexutils"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

//go:generate mockgen -source=state_hash_db.go -destination=./state_hash_db_mock.go -package=db

const StateHashPrefix = "dbh" // StateHashPrefix + 0x + blockNum(hex) -> StateHash

type StateHashDB interface {
	BaseDB

	PutStateHash(blockNumber uint64, stateHash []byte) error

	PutStateHashString(blockNumberHex string, stateHashString string) error

	GetStateHash(block int) (types.Hash, error)

	// GetFirstKey returns block number of first stateHash
	GetFirstKey() (uint64, error)

	// GetLastKey returns block number of last stateHash
	GetLastKey() (uint64, error)
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
	return &stateHashDB{&codeDB{backend: db}}
}

func MakeDefaultStateHashDBFromBaseDB(db BaseDB) StateHashDB {
	return &stateHashDB{&codeDB{backend: db.GetBackend()}}
}

// NewReadOnlyStateHashDB creates a new instance of read-only StateHashDB.
func NewReadOnlyStateHashDB(path string) (StateHashDB, error) {
	return newStateHashDB(path, &opt.Options{ReadOnly: true}, nil, nil)
}

func MakeStateHashDB(db *leveldb.DB, wo *opt.WriteOptions, ro *opt.ReadOptions) StateHashDB {
	return &stateHashDB{&codeDB{backend: db, wo: wo, ro: ro}}
}

func newStateHashDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (*stateHashDB, error) {
	base, err := newCodeDB(path, o, wo, ro)
	if err != nil {
		return nil, err
	}
	return &stateHashDB{base}, nil
}

type stateHashDB struct {
	BaseDB
}

func (db *stateHashDB) PutStateHash(blockNumber uint64, stateHash []byte) error {
	hex := strconv.FormatUint(blockNumber, 16)
	fullPrefix := StateHashPrefix + "0x" + hex
	err := db.Put([]byte(fullPrefix), stateHash)
	if err != nil {
		return fmt.Errorf("unable to put state hash for block %d: %v", blockNumber, err)
	}
	return nil
}

func (db *stateHashDB) PutStateHashString(blockNumberHex string, stateHash string) error {
	fullPrefix := StateHashPrefix + blockNumberHex
	err := db.Put([]byte(fullPrefix), hexutils.HexToBytes(strings.TrimPrefix(stateHash, "0x")))
	if err != nil {
		return fmt.Errorf("unable to put state hash for block %s: %v", blockNumberHex, err)
	}
	return nil
}

func (db *stateHashDB) GetStateHash(number int) (types.Hash, error) {
	hex := strconv.FormatUint(uint64(number), 16)
	stateRoot, err := db.Get([]byte(StateHashPrefix + "0x" + hex))
	if err != nil {
		return types.Hash{}, err
	}

	if stateRoot == nil {
		return types.Hash{}, nil
	}

	if len(stateRoot) != 32 {
		return types.Hash{}, fmt.Errorf("invalid state root length for block %d: expected 32 bytes, got %d bytes", number, len(stateRoot))
	}

	return types.BytesToHash(stateRoot), nil
}

// GetFirstKey returns the first block number for which we have a blockHash
func (db *stateHashDB) GetFirstKey() (uint64, error) {
	//iter := db.NewIterator([]byte(BlockHashPrefix), nil)
	//defer iter.Release()
	//
	//if !iter.Next() {
	//	return 0, fmt.Errorf("no blockHash found")
	//}
	//
	//firstBlock, err := DecodeStateHashDBKey(iter.Key())
	//if err != nil {
	//	return 0, err
	//}
	//return firstBlock, nil
	return 0, fmt.Errorf("not implemented")
}

// GetLastBlockHash returns the last block number for which we have a blockHash
func (db *stateHashDB) GetLastKey() (uint64, error) {
	//iter := db.NewIterator([]byte(BlockHashPrefix), nil)
	//defer iter.Release()
	//
	//if !iter.Last() {
	//	return 0, fmt.Errorf("no blockHash found")
	//}
	//
	//lastBlock, err := DecodeStateHashDBKey(iter.Key())
	//if err != nil {
	//	return 0, err
	//}
	//return lastBlock, nil
	return 0, fmt.Errorf("not implemented")
}

// StateHashKeyToUint64 converts a state hash key to a uint64
func StateHashKeyToUint64(hexBytes []byte) (uint64, error) {
	prefix := []byte(StateHashPrefix)

	if len(hexBytes) >= len(prefix) && bytes.HasPrefix(hexBytes, prefix) {
		hexBytes = hexBytes[len(prefix):]
	}

	res, err := strconv.ParseUint(string(hexBytes), 0, 64)

	if err != nil {
		return 0, fmt.Errorf("cannot parse uint %v; %v", string(hexBytes), err)
	}
	return res, nil
}
