package db

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/substate/protobuf"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/testutil"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

var testException = &substate.Exception{
	Block: 1001,
	Data: substate.ExceptionBlock{
		Transactions: map[int]substate.ExceptionTx{
			1: {
				PreTransaction:  &substate.WorldState{types.Address{0x01}: &substate.Account{Nonce: 1, Balance: uint256.NewInt(500)}},
				PostTransaction: &substate.WorldState{},
			},
		},
		PreBlock:  &substate.WorldState{},
		PostBlock: &substate.WorldState{},
	},
}

func TestExceptionDB_PutException_Nil(t *testing.T) {
	db := &exceptionDB{}
	err := db.PutException(nil)
	assert.Error(t, err)
}

func TestExceptionDB_PutAndGetException(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockBaseDB(ctrl)
	db := &exceptionDB{&codeDB{mockDB}}

	block := uint64(42)
	exc := &substate.Exception{
		Block: block,
		Data:  substate.ExceptionBlock{},
	}

	encoded, _ := protobuf.EncodeExceptionBlock(&exc.Data)
	mockDB.EXPECT().Put(ExceptionDBBlockPrefix(block), encoded).Return(nil)
	err := db.PutException(exc)
	assert.NoError(t, err)

	mockDB.EXPECT().Get(ExceptionDBBlockPrefix(block)).Return(encoded, nil)
	// decodeException will fail because pbExceptionData.Decode returns error (not implemented in test)
	_, err = db.GetException(block)
	assert.Error(t, err)
}

func TestExceptionDB_GetException_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockBaseDB(ctrl)
	db := &exceptionDB{&codeDB{mockDB}}

	block := uint64(100)
	mockDB.EXPECT().Get(ExceptionDBBlockPrefix(block)).Return(nil, leveldb.ErrNotFound)
	exc, err := db.GetException(block)
	assert.NoError(t, err)
	assert.Nil(t, exc)
}

func TestExceptionDB_GetException_EmptyData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockBaseDB(ctrl)
	db := &exceptionDB{&codeDB{mockDB}}

	block := uint64(101)
	mockDB.EXPECT().Get(ExceptionDBBlockPrefix(block)).Return([]byte{}, nil)
	exc, err := db.GetException(block)
	assert.Error(t, err)
	assert.Nil(t, exc)
}

func TestDecodeException_UnmarshalError(t *testing.T) {
	lookup := func(types.Hash) ([]byte, error) { return nil, nil }
	_, err := decodeException(lookup, 1, []byte{0x01, 0x02})
	assert.Error(t, err)
}

func TestDecodeException_Success(t *testing.T) {
	lookup := func(types.Hash) ([]byte, error) { return nil, nil }

	data := &protobuf.ExceptionBlock{
		Transactions: map[int32]*protobuf.ExceptionTx{
			1: {
				PreTransaction:  &protobuf.Alloc{Alloc: []*protobuf.AllocEntry{}},
				PostTransaction: &protobuf.Alloc{Alloc: []*protobuf.AllocEntry{}},
			},
		},
	}

	dataBytes, err := proto.Marshal(data)
	assert.NoError(t, err)

	exception, err := decodeException(lookup, 1, dataBytes)
	assert.NoError(t, err)
	assert.NotNil(t, exception)
	assert.Equal(t, uint64(1), exception.Block)
}

func TestExceptionDBBlockPrefix(t *testing.T) {
	block := uint64(123)
	prefix := ExceptionDBBlockPrefix(block)
	assert.True(t, len(prefix) > len(ExceptionDBPrefix))
}

func TestExceptionDb_GetFirstKeySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)
	kv := &testutil.KeyValue{}
	kv.PutU(ExceptionDBBlockPrefix(1), []byte{42})
	kv.PutU(ExceptionDBBlockPrefix(2), []byte{43})
	mockIter := iterator.NewArrayIterator(kv)

	mockDB.EXPECT().newIterator(gomock.Any()).Return(mockIter)

	db := &exceptionDB{mockDB}

	result, err := db.GetFirstKey()

	assert.Nil(t, err)
	assert.Equal(t, uint64(1), result)
}

func TestExceptionDb_GetFirstKeyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)

	// case 1 not found
	kv := &testutil.KeyValue{}
	mockIter := iterator.NewArrayIterator(kv)
	db := &exceptionDB{mockDB}

	mockDB.EXPECT().newIterator(gomock.Any()).Return(mockIter)

	result, err := db.GetFirstKey()

	assert.Equal(t, leveldb.ErrNotFound, err)
	assert.Equal(t, uint64(0), result)

	// case 2 decode error
	kv = &testutil.KeyValue{}
	kv.PutU([]byte{1, 2, 3}, []byte{42})
	mockIter = iterator.NewArrayIterator(kv)

	mockDB.EXPECT().newIterator(gomock.Any()).Return(mockIter)

	result, err = db.GetFirstKey()

	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), result)

}
func TestExceptionDb_GetLastKeySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)
	kv := &testutil.KeyValue{}
	kv.PutU(ExceptionDBBlockPrefix(1), []byte{30})
	kv.PutU(ExceptionDBBlockPrefix(5), []byte{42})
	mockIter := iterator.NewArrayIterator(kv)

	mockDB.EXPECT().newIterator(gomock.Any()).Return(mockIter)

	db := &exceptionDB{mockDB}

	result, err := db.GetLastKey()

	assert.Nil(t, err)
	assert.Equal(t, uint64(5), result)
}

func TestExceptionDb_GetLastKeyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)

	// case 1: no updateset found
	kv := &testutil.KeyValue{}
	mockIter := iterator.NewArrayIterator(kv)
	db := &exceptionDB{mockDB}

	mockDB.EXPECT().newIterator(gomock.Any()).Return(mockIter)

	result, err := db.GetLastKey()

	assert.Error(t, err)
	assert.Equal(t, leveldb.ErrNotFound, err)
	assert.Equal(t, uint64(0), result)

	// case 2: decode error
	kv = &testutil.KeyValue{}
	kv.PutU([]byte{1, 2, 3}, []byte{42})
	mockIter = iterator.NewArrayIterator(kv)

	mockDB.EXPECT().newIterator(gomock.Any()).Return(mockIter)

	result, err = db.GetLastKey()

	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), result)
}

func createDbAndPutException(dbPath string) (*exceptionDB, error) {
	db, err := newExceptionDB(dbPath, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot open db; %v", err)
	}

	err = db.PutException(testException)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestNewDefaultExceptionDB(t *testing.T) {
	dir := t.TempDir()
	db, err := NewDefaultExceptionDB(filepath.Join(dir, "testdb"))
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestNewExceptionDB(t *testing.T) {
	dir := t.TempDir()
	o := &opt.Options{}
	wo := &opt.WriteOptions{}
	ro := &opt.ReadOptions{}
	db, err := NewExceptionDB(filepath.Join(dir, "testdb2"), o, wo, ro)
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestMakeDefaultExceptionDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "ldb")
	ldb, err := leveldb.OpenFile(dbPath, nil)
	assert.NoError(t, err)
	exdb := MakeDefaultExceptionDB(ldb)
	assert.NotNil(t, exdb)
	err = ldb.Close()
	assert.NoError(t, err)
}

func TestMakeDefaultExceptionDBFromBaseDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "ldb3")
	ldb, err := leveldb.OpenFile(dbPath, nil)
	assert.NoError(t, err)

	baseDB := &baseDB{backend: ldb}
	exdb := MakeDefaultExceptionDBFromBaseDB(baseDB)
	assert.NotNil(t, exdb)
	err = ldb.Close()
	assert.NoError(t, err)
}

func TestNewReadOnlyExceptionDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "readonlydb")
	ldb, err := leveldb.OpenFile(dbPath, nil)
	assert.NoError(t, err)
	err = ldb.Close()
	assert.NoError(t, err)
	db, err := NewReadOnlyExceptionDB(dbPath)
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestMakeExceptionDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "ldb2")
	ldb, err := leveldb.OpenFile(dbPath, nil)
	assert.NoError(t, err)
	wo := &opt.WriteOptions{}
	ro := &opt.ReadOptions{}
	exdb := MakeExceptionDB(ldb, wo, ro)
	assert.NotNil(t, exdb)
	err = ldb.Close()
	assert.NoError(t, err)
}
