package db

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/0xsoniclabs/substate/protobuf"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"google.golang.org/protobuf/proto"
)

const ExceptionDBPrefix = "ex" // ExceptionDBPrefix + block (64-bit) -> ExceptionData

// ExceptionDB is a wrapper around CodeDB. It extends it with Has/Get/Put/DeleteSubstate functions.
//
//go:generate mockgen -source=exception_db.go -destination=./exception_db_mock.go -package=db
type ExceptionDB interface {
	BaseDB

	PutException(e *substate.Exception) error

	GetException(block uint64) (*substate.Exception, error)

	// GetFirstKey returns block number of first Exception. It returns an error if no Exception is found.
	GetFirstKey() (uint64, error)

	// GetLastKey returns block number of last Exception. It returns an error if no Exception is found.
	GetLastKey() (uint64, error)

	NewExceptionIterator(start int, numWorkers int) IIterator[*substate.Exception]

	decodeToException(data []byte, block uint64) (*substate.Exception, error)
}

// NewDefaultExceptionDB creates new instance of ExceptionDB with default options.
func NewDefaultExceptionDB(path string) (ExceptionDB, error) {
	return newExceptionDB(path, nil, nil, nil)
}

// NewExceptionDB creates new instance of ExceptionDB with customizable options.
// Note: Any of three options is nillable. If that's the case a default value for the option is set.
func NewExceptionDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (ExceptionDB, error) {
	return newExceptionDB(path, o, wo, ro)
}

func MakeDefaultExceptionDB(db *leveldb.DB) ExceptionDB {
	return &exceptionDB{&codeDB{backend: db}}
}

func MakeDefaultExceptionDBFromBaseDB(db BaseDB) ExceptionDB {
	return &exceptionDB{&codeDB{backend: db.GetBackend()}}
}

// NewReadOnlyExceptionDB creates a new instance of read-only ExceptionDB.
func NewReadOnlyExceptionDB(path string) (ExceptionDB, error) {
	return newExceptionDB(path, &opt.Options{ReadOnly: true}, nil, nil)
}

func MakeExceptionDB(db *leveldb.DB, wo *opt.WriteOptions, ro *opt.ReadOptions) ExceptionDB {
	return &exceptionDB{&codeDB{backend: db, wo: wo, ro: ro}}
}

func newExceptionDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (*exceptionDB, error) {
	base, err := newCodeDB(path, o, wo, ro)
	if err != nil {
		return nil, err
	}
	return &exceptionDB{base}, nil
}

type exceptionDB struct {
	CodeDB
}

func (db *exceptionDB) GetFirstKey() (uint64, error) {
	r := util.BytesPrefix([]byte(ExceptionDBPrefix))

	iter := db.newIterator(r)
	defer iter.Release()

	exist := iter.First()
	if exist {
		firstBlock, err := DecodeExceptionDBKey(iter.Key())
		if err != nil {
			return 0, fmt.Errorf("cannot decode exception key; %v", err)
		}
		return firstBlock, nil
	}
	return 0, leveldb.ErrNotFound
}

func (db *exceptionDB) GetLastKey() (uint64, error) {
	r := util.BytesPrefix([]byte(ExceptionDBPrefix))

	iter := db.newIterator(r)
	defer iter.Release()

	exist := iter.Last()
	if exist {
		lastBlock, err := DecodeExceptionDBKey(iter.Key())
		if err != nil {
			return 0, fmt.Errorf("cannot decode exception key; %v", err)
		}
		return lastBlock, nil
	}
	return 0, leveldb.ErrNotFound
}

func (db *exceptionDB) PutException(e *substate.Exception) error {
	if e == nil {
		return errors.New("cannot put nil exception")
	}

	value, err := protobuf.EncodeExceptionBlock(&e.Data)
	if err != nil {
		return fmt.Errorf("cannot encode exception data for block %v; %w", e.Block, err)
	}
	return db.Put(ExceptionDBBlockPrefix(e.Block), value)
}

// GetException retrieves exception for a given block number.
func (db *exceptionDB) GetException(block uint64) (*substate.Exception, error) {
	data, err := db.Get(ExceptionDBBlockPrefix(block))
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("cannot get exception for block %d; %w", block, err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("exception data for block %d is empty", block)
	}

	return decodeException(db.GetCode, block, data)
}

// NewExceptionIterator returns iterator which iterates over Exceptions.
func (db *exceptionDB) NewExceptionIterator(start int, numWorkers int) IIterator[*substate.Exception] {
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	iter := newExceptionIterator(db, blockTx)

	iter.start(numWorkers)

	return iter
}

func (db *exceptionDB) decodeToException(data []byte, block uint64) (*substate.Exception, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("exception data for block %d is empty", block)
	}

	exceptionBlock, err := decodeException(db.GetCode, block, data)
	if err != nil {
		return nil, fmt.Errorf("cannot decode exception data for block %v; %w", block, err)
	}

	return exceptionBlock, nil
}

func decodeException(lookup func(types.Hash) ([]byte, error), block uint64, data []byte) (*substate.Exception, error) {
	pbExceptionData := &protobuf.ExceptionBlock{}
	if err := proto.Unmarshal(data, pbExceptionData); err != nil {
		return nil, fmt.Errorf("cannot decode exception data from protobuf block: %v, %w", block, err)
	}
	exceptionBlock, err := pbExceptionData.Decode(lookup)
	if err != nil {
		return nil, fmt.Errorf("cannot decode exception data for block %v; %w", block, err)
	}

	if exceptionBlock == nil {
		return nil, fmt.Errorf("decoded exception data for block %v is nil", block)
	}

	return &substate.Exception{
		Block: block,
		Data:  *exceptionBlock,
	}, nil
}

// ExceptionDBBlockPrefix returns ExceptionDBPrefix with appended
// block number creating prefix used in baseDB for Substates.
func ExceptionDBBlockPrefix(block uint64) []byte {
	return append([]byte(ExceptionDBPrefix), BlockToBytes(block)...)
}

func DecodeExceptionDBKey(key []byte) (block uint64, err error) {
	prefix := ExceptionDBPrefix
	if len(key) != len(prefix)+8 {
		err = fmt.Errorf("invalid length of exception key: %v", len(key))
		return
	}
	if p := string(key[:len(prefix)]); p != prefix {
		err = fmt.Errorf("invalid prefix of exception key: %#x", p)
		return
	}
	blockTx := key[len(prefix):]
	block = binary.BigEndian.Uint64(blockTx[0:8])
	return
}
