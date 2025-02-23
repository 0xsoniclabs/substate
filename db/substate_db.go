package db

import (
	"encoding/binary"
	"fmt"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/urfave/cli/v2"
)

const SubstateDBPrefix = "1s" // SubstateDBPrefix + block (64-bit) + tx (64-bit) -> substateRLP

// ISubstateDB is a wrapper around CodeDB. It extends it with Has/Get/Put/DeleteSubstate functions.
//
//go:generate mockgen -source=substate_db.go -destination=./substate_db_mock.go -package=db
type ISubstateDB interface {
	CodeDB

	// HasSubstate returns true if the DB does contain Substate for given block and tx number.
	HasSubstate(block uint64, tx int) (bool, error)

	// GetSubstate gets the Substate for given block and tx number.
	GetSubstate(block uint64, tx int) (*substate.Substate, error)

	// GetBlockSubstates returns all existing substates for given block.
	GetBlockSubstates(block uint64) (map[int]*substate.Substate, error)

	// PutSubstate inserts given substate to DB.
	PutSubstate(substate *substate.Substate) error

	// DeleteSubstate deletes Substate for given block and tx number.
	DeleteSubstate(block uint64, tx int) error

	NewSubstateIterator(start int, numWorkers int) IIterator[*substate.Substate]

	NewSubstateTaskPool(name string, taskFunc SubstateTaskFunc, first, last uint64, ctx *cli.Context) *SubstateTaskPool

	// GetFirstSubstate returns last substate (block and transaction wise) inside given DB.
	GetFirstSubstate() *substate.Substate

	// GetLastSubstate returns last substate (block and transaction wise) inside given DB.
	GetLastSubstate() (*substate.Substate, error)

	// SetSubstateEncoding sets the decoder func to the provided encoding
	SetSubstateEncoding(encoding string) (*SubstateDB, error)

	// GetSubstateEncoding returns the currently configured encoding
	GetSubstateEncoding() string
}

// NewDefaultSubstateDB creates new instance of ISubstateDB with default options.
func NewDefaultSubstateDB(path string) (ISubstateDB, error) {
	return newSubstateDB(path, nil, nil, nil)
}

// NewSubstateDB creates new instance of ISubstateDB with customizable options.
// Note: Any of three options is nillable. If that's the case a default value for the option is set.
func NewSubstateDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (ISubstateDB, error) {
	return newSubstateDB(path, o, wo, ro)
}

func MakeDefaultSubstateDB(db *leveldb.DB) ISubstateDB {
	sdb := &SubstateDB{&codeDB{&baseDB{backend: db}}, nil}
	sdb, _ = sdb.SetSubstateEncoding("default")
	return sdb
}

func MakeDefaultSubstateDBFromBaseDB(db BaseDB) ISubstateDB {
	sdb := &SubstateDB{&codeDB{&baseDB{backend: db.getBackend()}}, nil}
	sdb, _ = sdb.SetSubstateEncoding("default")
	return sdb
}

// NewReadOnlySubstateDB creates a new instance of read-only ISubstateDB.
func NewReadOnlySubstateDB(path string) (ISubstateDB, error) {
	return newSubstateDB(path, &opt.Options{ReadOnly: true}, nil, nil)
}

func MakeSubstateDB(db *leveldb.DB, wo *opt.WriteOptions, ro *opt.ReadOptions) ISubstateDB {
	sdb := &SubstateDB{&codeDB{&baseDB{backend: db, wo: wo, ro: ro}}, nil}
	sdb, _ = sdb.SetSubstateEncoding("default")
	return sdb
}

func newSubstateDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (*SubstateDB, error) {
	base, err := newCodeDB(path, o, wo, ro)
	if err != nil {
		return nil, err
	}

	sdb := &SubstateDB{base, nil}
	sdb, _ = sdb.SetSubstateEncoding("default")
	return sdb, nil
}

type SubstateDB struct {
	*codeDB
	encoding *substateEncoding
}

func (db *SubstateDB) GetFirstSubstate() *substate.Substate {
	iter := db.NewSubstateIterator(0, 1)

	defer iter.Release()

	if iter.Next() {
		return iter.Value()
	}

	return nil
}

func (db *SubstateDB) HasSubstate(block uint64, tx int) (bool, error) {
	return db.Has(SubstateDBKey(block, tx))
}

// GetSubstate returns substate for given block and tx number if exists within DB.
func (db *SubstateDB) GetSubstate(block uint64, tx int) (*substate.Substate, error) {
	val, err := db.Get(SubstateDBKey(block, tx))
	if err != nil {
		return nil, fmt.Errorf("cannot get substate block: %v, tx: %v from db; %w", block, tx, err)
	}

	return db.decodeToSubstate(val, block, tx)
}

// GetBlockSubstates returns substates for given block if exists within DB.
func (db *SubstateDB) GetBlockSubstates(block uint64) (map[int]*substate.Substate, error) {
	var err error

	txSubstate := make(map[int]*substate.Substate)

	prefix := SubstateDBBlockPrefix(block)

	iter := db.backend.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		b, tx, err := DecodeSubstateDBKey(key)
		if err != nil {
			return nil, fmt.Errorf("record-replay: invalid substate key found for block %v: %w", block, err)
		}

		if block != b {
			return nil, fmt.Errorf("record-replay: GetBlockSubstates(%v) iterated substates from block %v", block, b)
		}

		sbstt, err := db.decodeToSubstate(value, block, tx)
		if err != nil {
			return nil, fmt.Errorf("failed to decode substate, block %v, tx: %v; %w", block, tx, err)
		}

		txSubstate[tx] = sbstt
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		return nil, err
	}

	return txSubstate, nil
}

func (db *SubstateDB) PutSubstate(ss *substate.Substate) error {
	for i, account := range ss.InputSubstate {
		err := db.PutCode(account.Code)
		if err != nil {
			return fmt.Errorf("cannot put preState code from substate-account %v block %v, %v tx into db; %w", i, ss.Block, ss.Transaction, err)
		}
	}

	for i, account := range ss.OutputSubstate {
		err := db.PutCode(account.Code)
		if err != nil {
			return fmt.Errorf("cannot put postState code from substate-account %v block %v, %v tx into db; %w", i, ss.Block, ss.Transaction, err)
		}
	}

	if msg := ss.Message; msg.To == nil {
		err := db.PutCode(msg.Data)
		if err != nil {
			return fmt.Errorf("cannot put input data from substate block %v, %v tx into db; %v", ss.Block, ss.Transaction, err)
		}
	}

	key := SubstateDBKey(ss.Block, ss.Transaction)

	value, err := db.encodeSubstate(ss, ss.Block, ss.Transaction)
	if err != nil {
		return fmt.Errorf("cannot encode substate block %v, tx %v; %v", ss.Block, ss.Transaction, err)
	}

	return db.Put(key, value)
}

func (db *SubstateDB) DeleteSubstate(block uint64, tx int) error {
	return db.Delete(SubstateDBKey(block, tx))
}

// NewSubstateIterator returns iterator which iterates over Substates.
func (db *SubstateDB) NewSubstateIterator(start int, numWorkers int) IIterator[*substate.Substate] {
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	iter := newSubstateIterator(db, blockTx)

	iter.start(numWorkers)

	return iter
}

func (db *SubstateDB) NewSubstateTaskPool(name string, taskFunc SubstateTaskFunc, first, last uint64, ctx *cli.Context) *SubstateTaskPool {
	return &SubstateTaskPool{
		Name:     name,
		TaskFunc: taskFunc,

		First: first,
		Last:  last,

		Workers:         ctx.Int(WorkersFlag.Name),
		SkipTransferTxs: ctx.Bool(SkipTransferTxsFlag.Name),
		SkipCallTxs:     ctx.Bool(SkipCallTxsFlag.Name),
		SkipCreateTxs:   ctx.Bool(SkipCreateTxsFlag.Name),

		Ctx: ctx,

		DB: db,
	}
}

// getLongestEncodedKeyZeroPrefixLength returns longest index of biggest block number to be search for in its search
func (db *SubstateDB) getLongestEncodedKeyZeroPrefixLength() (byte, error) {
	var i byte
	for i = 0; i < 8; i++ {
		startingIndex := make([]byte, 8)
		startingIndex[i] = 1
		if db.hasKeyValuesFor([]byte(SubstateDBPrefix), startingIndex) {
			return i, nil
		}
	}

	return 0, fmt.Errorf("unable to find prefix of substate with biggest block")
}

// getLastBlock returns block number of last substate
func (db *SubstateDB) getLastBlock() (uint64, error) {
	zeroBytes, err := db.getLongestEncodedKeyZeroPrefixLength()
	if err != nil {
		return 0, err
	}

	var lastKeyPrefix []byte
	if zeroBytes > 0 {
		blockBytes := make([]byte, zeroBytes)

		lastKeyPrefix = append([]byte(SubstateDBPrefix), blockBytes...)
	} else {
		lastKeyPrefix = []byte(SubstateDBPrefix)
	}

	substatePrefixSize := len([]byte(SubstateDBPrefix))

	// binary search for biggest key
	for {
		nextBiggestPrefixValue, err := db.binarySearchForLastPrefixKey(lastKeyPrefix)
		if err != nil {
			return 0, err
		}
		lastKeyPrefix = append(lastKeyPrefix, nextBiggestPrefixValue)
		// we have all 8 bytes of uint64 encoded block
		if len(lastKeyPrefix) == (substatePrefixSize + 8) {
			// full key is already found
			substateBlockValue := lastKeyPrefix[substatePrefixSize:]

			if len(substateBlockValue) != 8 {
				return 0, fmt.Errorf("undefined behaviour in GetLastSubstate search; retrieved block bytes can't be converted")
			}
			return binary.BigEndian.Uint64(substateBlockValue), nil
		}
	}
}

func (db *SubstateDB) GetLastSubstate() (*substate.Substate, error) {
	block, err := db.getLastBlock()
	if err != nil {
		return nil, err
	}
	substates, err := db.GetBlockSubstates(block)
	if err != nil {
		return nil, fmt.Errorf("cannot get block substates; %w", err)
	}
	if len(substates) == 0 {
		return nil, fmt.Errorf("block %v doesn't have any substates", block)
	}
	maxTx := 0
	for txIdx := range substates {
		if txIdx > maxTx {
			maxTx = txIdx
		}
	}
	return substates[maxTx], nil
}

// BlockToBytes returns binary BigEndian representation of given block number.
func BlockToBytes(block uint64) []byte {
	blockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(blockBytes[0:8], block)
	return blockBytes
}

// SubstateDBKey returns SubstateDBPrefix with appended
// block number creating key used in baseDB for Substates.
func SubstateDBKey(block uint64, tx int) []byte {
	prefix := []byte(SubstateDBPrefix)

	blockTx := make([]byte, 16)
	binary.BigEndian.PutUint64(blockTx[0:8], block)
	binary.BigEndian.PutUint64(blockTx[8:16], uint64(tx))

	return append(prefix, blockTx...)
}

// SubstateDBBlockPrefix returns SubstateDBPrefix with appended
// block number creating prefix used in baseDB for Substates.
func SubstateDBBlockPrefix(block uint64) []byte {
	return append([]byte(SubstateDBPrefix), BlockToBytes(block)...)
}

// DecodeSubstateDBKey decodes key created by SubstateDBBlockPrefix back to block and tx number.
func DecodeSubstateDBKey(key []byte) (block uint64, tx int, err error) {
	prefix := SubstateDBPrefix
	if len(key) != len(prefix)+8+8 {
		err = fmt.Errorf("invalid length of substate db key: %v", len(key))
		return
	}
	if p := string(key[:len(prefix)]); p != prefix {
		err = fmt.Errorf("invalid prefix of substate db key: %#x", p)
		return
	}
	blockTx := key[len(prefix):]
	block = binary.BigEndian.Uint64(blockTx[0:8])
	tx = int(binary.BigEndian.Uint64(blockTx[8:16]))
	return
}
