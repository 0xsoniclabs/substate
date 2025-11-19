package db

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/0xsoniclabs/substate/types"
	"github.com/status-im/keycard-go/hexutils"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

//go:generate mockgen -source=block_hash_db.go -destination=./block_hash_db_mock.go -package=db

const BlockHashPrefix = "bh"

type BlockHashDB interface {
	BaseDB

	PutBlockHash(blockNumber uint64, blockHash []byte) error

	PutBlockHashString(blockNumberHex string, blockHashString string) error

	GetBlockHash(block int) (types.Hash, error)

	// GetFirstKey returns block number of first BlockHash
	GetFirstKey() (uint64, error)

	// GetLastKey returns block number of last BlockHash
	GetLastKey() (uint64, error)
}

// NewDefaultBlockHashDB creates new instance of BlockHashDB with default options.
func NewDefaultBlockHashDB(path string) (BlockHashDB, error) {
	return newBlockHashDB(path, nil, nil, nil)
}

// NewBlockHashDB creates new instance of BlockHashDB with customizable options.
// Note: Any of three options is nillable. If that's the case a default value for the option is set.
func NewBlockHashDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (BlockHashDB, error) {
	return newBlockHashDB(path, o, wo, ro)
}

func MakeDefaultBlockHashDB(db *leveldb.DB) BlockHashDB {
	return &blockHashDB{&codeDB{backend: db}}
}

func MakeDefaultBlockHashDBFromBaseDB(db BaseDB) BlockHashDB {
	return &blockHashDB{&codeDB{backend: db.GetBackend()}}
}

// NewReadOnlyBlockHashDB creates a new instance of read-only BlockHashDB.
func NewReadOnlyBlockHashDB(path string) (BlockHashDB, error) {
	return newBlockHashDB(path, &opt.Options{ReadOnly: true}, nil, nil)
}

func MakeBlockHashDB(db *leveldb.DB, wo *opt.WriteOptions, ro *opt.ReadOptions) BlockHashDB {
	return &blockHashDB{&codeDB{backend: db, wo: wo, ro: ro}}
}

func newBlockHashDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (*blockHashDB, error) {
	base, err := newCodeDB(path, o, wo, ro)
	if err != nil {
		return nil, err
	}
	return &blockHashDB{base}, nil
}

type blockHashDB struct {
	BaseDB
}

func (db *blockHashDB) PutBlockHash(blockNumber uint64, blockHash []byte) error {
	blockNumberHex := "0x" + strconv.FormatUint(blockNumber, 16)
	blockHashString := fmt.Sprintf("0x%x", blockHash)
	return db.PutBlockHashString(blockNumberHex, blockHashString)
}

func (db *blockHashDB) PutBlockHashString(blockNumberHex string, blockHashString string) error {
	bn, err := strconv.ParseUint(strings.TrimPrefix(blockNumberHex, "0x"), 16, 64)
	if err != nil {
		return fmt.Errorf("invalid block number %s: %v", blockNumberHex, err)
	}
	fullPrefix := BlockHashDBKey(bn)
	err = db.Put(fullPrefix, hexutils.HexToBytes(strings.TrimPrefix(blockHashString, "0x")))
	if err != nil {
		return fmt.Errorf("unable to put blockHash for block %s: %v", blockNumberHex, err)
	}
	return nil
}

func (db *blockHashDB) GetBlockHash(number int) (types.Hash, error) {
	blockHash, err := db.Get(BlockHashDBKey(uint64(number)))
	if err != nil {
		return types.Hash{}, err
	}

	if blockHash == nil {
		return types.Hash{}, nil
	}

	if len(blockHash) != 32 {
		return types.Hash{}, fmt.Errorf("invalid blockHash length for block %d: expected 32 bytes, got %d bytes", number, len(blockHash))
	}

	return types.Hash(blockHash), nil
}

// GetFirstKey returns the first block number for which we have a blockHash
func (db *blockHashDB) GetFirstKey() (uint64, error) {
	iter := db.NewIterator([]byte(BlockHashPrefix), nil)
	defer iter.Release()

	if !iter.Next() {
		return 0, fmt.Errorf("no blockHash found")
	}

	firstBlock, err := DecodeBlockHashDBKey(iter.Key())
	if err != nil {
		return 0, err
	}
	return firstBlock, nil
}

// GetLastBlockHash returns the last block number for which we have a blockHash
func (db *blockHashDB) GetLastKey() (uint64, error) {
	iter := db.NewIterator([]byte(BlockHashPrefix), nil)
	defer iter.Release()

	if !iter.Last() {
		return 0, fmt.Errorf("no blockHash found")
	}

	lastBlock, err := DecodeBlockHashDBKey(iter.Key())
	if err != nil {
		return 0, err
	}
	return lastBlock, nil
}

func BlockHashDBKey(block uint64) []byte {
	prefix := []byte(BlockHashPrefix)
	blockByte := make([]byte, 8)
	binary.BigEndian.PutUint64(blockByte[0:8], block)
	return append(prefix, blockByte...)
}

// DecodeBlockHashDBKey decodes a blockHash key into a block number
func DecodeBlockHashDBKey(data []byte) (uint64, error) {
	if len(data) < len(BlockHashPrefix)+8 {
		return 0, fmt.Errorf("invalid length of blockHash key, expected at least %d, got %d", len(BlockHashPrefix)+8, len(data))
	}
	if !bytes.HasPrefix(data, []byte(BlockHashPrefix)) {
		return 0, fmt.Errorf("invalid prefix of blockHash key")
	}
	block := binary.BigEndian.Uint64(data[len(BlockHashPrefix):])
	return block, nil
}
