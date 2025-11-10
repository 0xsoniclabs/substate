// Copyright 2025 Sonic Labs
// This file is part of Aida Testing Infrastructure for Sonic
//
// Aida is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Aida is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Aida. If not, see <http://www.gnu.org/licenses/>.

package db

//go:generate mockgen -source hash_provider.go -destination hash_provider_mock.go -package db

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/0xsoniclabs/substate/types"
	"github.com/status-im/keycard-go/hexutils"
)

const (
	StateRootHashPrefix = "dbh"
	BlockHashPrefix     = "bh"
)

// ClientInterface defines the methods that an RPC client must implement.
type IRpcClient interface {
	Call(result interface{}, method string, args ...interface{}) error
}

type HashProvider interface {
	GetStateRootHash(blockNumber int) (types.Hash, error)
	GetBlockHash(blockNumber int) (types.Hash, error)
}

func MakeHashProvider(db BaseDB) HashProvider {
	return &hashProvider{db}
}

type hashProvider struct {
	db BaseDB
}

func (p *hashProvider) GetBlockHash(number int) (types.Hash, error) {
	blockHash, err := p.db.Get(BlockHashDBKey(uint64(number)))
	if err != nil {
		return types.Hash{}, err
	}

	if blockHash == nil {
		return types.Hash{}, nil
	}

	if len(blockHash) != 32 {
		return types.Hash{}, fmt.Errorf("invalid block hash length for block %d: expected 32 bytes, got %d bytes", number, len(blockHash))
	}

	return types.Hash(blockHash), nil
}

func (p *hashProvider) GetStateRootHash(number int) (types.Hash, error) {
	hex := strconv.FormatUint(uint64(number), 16)
	stateRoot, err := p.db.Get([]byte(StateRootHashPrefix + "0x" + hex))
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

// SaveStateRoot saves the state root hash to the database
func SaveStateRoot(db BaseDB, blockNumber string, stateRoot string) error {
	fullPrefix := StateRootHashPrefix + blockNumber
	err := db.Put([]byte(fullPrefix), hexutils.HexToBytes(strings.TrimPrefix(stateRoot, "0x")))
	if err != nil {
		return fmt.Errorf("unable to put state hash for block %s: %v", blockNumber, err)
	}
	return nil
}

// SaveBlockHash saves the block hash to the database
func SaveBlockHash(db BaseDB, blockNumber string, hash string) error {
	bn, err := strconv.ParseUint(strings.TrimPrefix(blockNumber, "0x"), 16, 64)
	if err != nil {
		return fmt.Errorf("invalid block number %s: %v", blockNumber, err)
	}
	fullPrefix := BlockHashDBKey(bn)
	err = db.Put(fullPrefix, hexutils.HexToBytes(strings.TrimPrefix(hash, "0x")))
	if err != nil {
		return fmt.Errorf("unable to put state hash for block %s: %v", blockNumber, err)
	}
	return nil
}

// getBlockByNumber get block from the rpc node
func GetBlockByNumber(client IRpcClient, blockNumber string) (map[string]interface{}, error) {
	var block map[string]interface{}
	err := client.Call(&block, "eth_getBlockByNumber", blockNumber, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get block %s: %v", blockNumber, err)
	}
	return block, nil
}

// StateHashKeyToUint64 converts a state hash key to a uint64
func StateHashKeyToUint64(hexBytes []byte) (uint64, error) {
	prefix := []byte(StateRootHashPrefix)

	if len(hexBytes) >= len(prefix) && bytes.HasPrefix(hexBytes, prefix) {
		hexBytes = hexBytes[len(prefix):]
	}

	res, err := strconv.ParseUint(string(hexBytes), 0, 64)

	if err != nil {
		return 0, fmt.Errorf("cannot parse uint %v; %v", string(hexBytes), err)
	}
	return res, nil
}

// GetFirstStateHash returns the first block number for which we have a state hash
func GetFirstStateHash(db BaseDB) (uint64, error) {
	// TODO MATEJ will be fixed in future commit
	//iter := db.NewIterator([]byte(StateRootHashPrefix), []byte("0x"))
	//
	//defer iter.Release()
	//
	//// start with writing first block
	//if !iter.Next() {
	//	return 0, fmt.Errorf("no state hash found")
	//}
	//
	//firstStateHashBlock, err := StateHashKeyToUint64(iter.Key())
	//if err != nil {
	//	return 0, err
	//}
	//return firstStateHashBlock, nil
	return 0, fmt.Errorf("not implemented")
}

// GetLastStateHash returns the last block number for which we have a state hash
func GetLastStateHash(db BaseDB) (uint64, error) {
	// TODO MATEJ will be fixed in future commit
	//return GetLastKey(db, StateRootHashPrefix)
	return 0, fmt.Errorf("not implemented")
}

// GetFirstBlockHash returns the first block number for which we have a block hash
func GetFirstBlockHash(db BaseDB) (uint64, error) {
	iter := db.NewIterator([]byte(BlockHashPrefix), nil)
	defer iter.Release()

	if !iter.Next() {
		return 0, fmt.Errorf("no block hash found")
	}

	firstBlock, err := DecodeBlockHashDBKey(iter.Key())
	if err != nil {
		return 0, err
	}
	return firstBlock, nil
}

// GetLastBlockHash returns the last block number for which we have a block hash
func GetLastBlockHash(db BaseDB) (uint64, error) {
	iter := db.NewIterator([]byte(BlockHashPrefix), nil)
	defer iter.Release()

	if !iter.Last() {
		return 0, fmt.Errorf("no block hash found")
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

// DecodeBlockHashDBKey decodes a block hash key into a block number
func DecodeBlockHashDBKey(data []byte) (uint64, error) {
	if len(data) < len(BlockHashPrefix)+8 {
		return 0, fmt.Errorf("invalid length of block hash key, expected at least %d, got %d", len(BlockHashPrefix)+8, len(data))
	}
	if !bytes.HasPrefix(data, []byte(BlockHashPrefix)) {
		return 0, fmt.Errorf("invalid prefix of block hash key")
	}
	block := binary.BigEndian.Uint64(data[len(BlockHashPrefix):])
	return block, nil
}
