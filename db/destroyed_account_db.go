package db

import (
	"encoding/binary"
	"fmt"

	"github.com/0xsoniclabs/substate/types"
	"github.com/syndtr/goleveldb/leveldb"
	ldbiterator "github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	DestroyedAccountPrefix = "da" // DestroyedAccountPrefix + block (64-bit) -> SuicidedAccountLists
)

//go:generate mockgen -source=destroyed_account_db.go -destination=./destroyed_account_db_mock.go -package=db
type DestroyedAccountDB interface {
	BaseDB

	// SetSubstateEncoding sets the runtime encoding/decoding
	SetSubstateEncoding(schema SubstateEncodingSchema) error

	// GetSubstateEncoding returns the encoding schema in use.
	GetSubstateEncoding() SubstateEncodingSchema

	// Set the accounts that were destroyed in a specific block and transaction
	SetDestroyedAccounts(block uint64, tx int, destroyed []types.Address, resurrected []types.Address) error

	// Get the accounts that were destroyed in a specific block and transaction
	GetDestroyedAccounts(block uint64, tx int) ([]types.Address, []types.Address, error)

	// Get the accounts that were destroyed in a specific block range
	GetAccountsDestroyedInRange(from, to uint64) ([]types.Address, error)

	// Get the first destroyed account key
	GetFirstKey() (uint64, error)

	// Get the last destroyed account key
	GetLastKey() (uint64, error)
}

func NewDefaultDestroyedAccountDB(destroyedAccountDir string) (DestroyedAccountDB, error) {
	return newDestroyedAccountDB(destroyedAccountDir, nil, nil, nil)
}

func MakeDefaultDestroyedAccountDBFromBaseDB(db BaseDB) (DestroyedAccountDB, error) {
	value, err := MakeDefaultDestroyedAccountDBFromBaseDBWithEncoding(db, DefaultEncodingSchema)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func MakeDefaultDestroyedAccountDBFromBaseDBWithEncoding(db BaseDB, schema SubstateEncodingSchema) (DestroyedAccountDB, error) {
	encoding, err := newDestroyedAccountEncoding(schema)
	if err != nil {
		return nil, err
	}
	return &destroyedAccountDB{
		db.GetBackend(),
		nil,
		nil,
		*encoding,
	}, nil
}

func NewReadOnlyDestroyedAccountDB(destroyedAccountDir string) (DestroyedAccountDB, error) {
	return newDestroyedAccountDB(destroyedAccountDir, &opt.Options{ReadOnly: true}, nil, nil)
}

func newDestroyedAccountDB(destroyedAccountDir string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (DestroyedAccountDB, error) {
	backend, err := leveldb.OpenFile(destroyedAccountDir, o)
	if err != nil {
		return nil, fmt.Errorf("error opening deletion-db %s: %w", destroyedAccountDir, err)
	}
	encoding, err := newDestroyedAccountEncoding(DefaultEncodingSchema)
	if err != nil {
		return nil, err
	}
	return &destroyedAccountDB{
		backend:  backend,
		wo:       wo,
		ro:       ro,
		encoding: *encoding,
	}, nil
}

type destroyedAccountDB struct {
	backend  DbAdapter
	wo       *opt.WriteOptions
	ro       *opt.ReadOptions
	encoding destroyedAccountEncoding
}

// SuicidedAccountLists is value structure which represents the list of accounts
// that were either suicided and resurrected.
type SuicidedAccountLists struct {
	DestroyedAccounts   []types.Address
	ResurrectedAccounts []types.Address
}

func (db *destroyedAccountDB) SetDestroyedAccounts(block uint64, tx int, des []types.Address, res []types.Address) error {
	accountList := SuicidedAccountLists{DestroyedAccounts: des, ResurrectedAccounts: res}
	value, err := db.encoding.encode(accountList)
	if err != nil {
		return err
	}
	return db.Put(EncodeDestroyedAccountKey(block, tx), value)
}

func (db *destroyedAccountDB) GetDestroyedAccounts(block uint64, tx int) ([]types.Address, []types.Address, error) {
	data, err := db.Get(EncodeDestroyedAccountKey(block, tx))
	if data == nil {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	list, err := db.encoding.decode(data)
	if err != nil {
		return nil, nil, err
	}
	return list.DestroyedAccounts, list.ResurrectedAccounts, err
}

// GetAccountsDestroyedInRange get list of all accounts between block from and to (including from and to).
func (db *destroyedAccountDB) GetAccountsDestroyedInRange(from, to uint64) ([]types.Address, error) {
	startingBlockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(startingBlockBytes, from)

	iter := db.NewIterator([]byte(DestroyedAccountPrefix), startingBlockBytes)
	defer iter.Release()
	isDestroyed := make(map[types.Address]bool)
	for iter.Next() {
		block, _, err := DecodeDestroyedAccountKey(iter.Key())
		if err != nil {
			return nil, err
		}
		if block > to {
			break
		}
		list, err := db.encoding.decode(iter.Value())
		if err != nil {
			return nil, err
		}
		for _, addr := range list.DestroyedAccounts {
			isDestroyed[addr] = true
		}
		for _, addr := range list.ResurrectedAccounts {
			isDestroyed[addr] = false
		}
	}

	var accountList []types.Address
	for addr, isDeleted := range isDestroyed {
		if isDeleted {
			accountList = append(accountList, addr)
		}
	}
	return accountList, nil
}

// GetFirstKey returns the first block number in the database.
func (db *destroyedAccountDB) GetFirstKey() (uint64, error) {
	iter := db.NewIterator([]byte(DestroyedAccountPrefix), nil)
	defer iter.Release()

	exist := iter.Next()
	if exist {
		firstBlock, _, err := DecodeDestroyedAccountKey(iter.Key())
		if err != nil {
			return 0, fmt.Errorf("cannot decode updateset key; %w", err)
		}
		return firstBlock, nil
	}
	return 0, leveldb.ErrNotFound
}

// GetLastKey returns the last block number in the database.
// if not found then returns 0
func (db *destroyedAccountDB) GetLastKey() (uint64, error) {
	var block uint64
	var err error
	iter := db.NewIterator([]byte(DestroyedAccountPrefix), nil)
	defer iter.Release()

	exist := iter.Last()
	if exist {
		block, _, err = DecodeDestroyedAccountKey(iter.Key())
		if err != nil {
			return 0, fmt.Errorf("cannot decode updateset key; %w", err)
		}
		return block, nil
	}
	return 0, leveldb.ErrNotFound
}

func (db *destroyedAccountDB) stats(stats *leveldb.DBStats) error {
	return db.backend.Stats(stats)
}

func (db *destroyedAccountDB) GetBackend() DbAdapter {
	return db.backend
}

func (db *destroyedAccountDB) Put(key []byte, value []byte) error {
	return db.backend.Put(key, value, db.wo)
}

func (db *destroyedAccountDB) Delete(key []byte) error {
	return db.backend.Delete(key, db.wo)
}

func (db *destroyedAccountDB) Close() error {
	return db.backend.Close()
}

func (db *destroyedAccountDB) Has(key []byte) (bool, error) {
	return db.backend.Has(key, db.ro)
}

func (db *destroyedAccountDB) Get(key []byte) ([]byte, error) {
	return db.backend.Get(key, db.ro)
}

func (db *destroyedAccountDB) NewBatch() Batch {
	return newBatch(db.backend)
}

func (db *destroyedAccountDB) NewIterator(prefix []byte, start []byte) ldbiterator.Iterator {
	r := util.BytesPrefix(prefix)
	r.Start = append(r.Start, start...)
	return db.backend.NewIterator(r, db.ro)
}

func (db *destroyedAccountDB) newIterator(r *util.Range) ldbiterator.Iterator {
	return db.backend.NewIterator(r, db.ro)
}

func (db *destroyedAccountDB) Stat(property string) (string, error) {
	return db.backend.GetProperty(property)
}

func (db *destroyedAccountDB) Compact(start []byte, limit []byte) error {
	return db.backend.CompactRange(util.Range{Start: start, Limit: limit})
}

func (db *destroyedAccountDB) hasKeyValuesFor(prefix []byte, start []byte) bool {
	iter := db.NewIterator(prefix, start)
	defer iter.Release()
	return iter.Next()
}

func (db *destroyedAccountDB) binarySearchForLastPrefixKey(lastKeyPrefix []byte) (byte, error) {
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

func EncodeDestroyedAccountKey(block uint64, tx int) []byte {
	prefix := []byte(DestroyedAccountPrefix)
	key := make([]byte, len(prefix)+12)
	copy(key[0:], prefix)
	binary.BigEndian.PutUint64(key[len(prefix):], block)
	binary.BigEndian.PutUint32(key[len(prefix)+8:], uint32(tx))
	return key
}

func DecodeDestroyedAccountKey(data []byte) (uint64, int, error) {
	if len(data) != len(DestroyedAccountPrefix)+12 {
		return 0, 0, fmt.Errorf("invalid length of destroyed account key, expected %d, got %d", len(DestroyedAccountPrefix)+12, len(data))
	}
	if string(data[0:len(DestroyedAccountPrefix)]) != DestroyedAccountPrefix {
		return 0, 0, fmt.Errorf("invalid prefix of destroyed account key")
	}
	block := binary.BigEndian.Uint64(data[len(DestroyedAccountPrefix):])
	tx := binary.BigEndian.Uint32(data[len(DestroyedAccountPrefix)+8:])
	return block, int(tx), nil
}
