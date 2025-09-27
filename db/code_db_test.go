package db

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/0xsoniclabs/substate/types/hash"
)

var testCode = []byte{1}

func TestCodeDB_PutCode(t *testing.T) {
	dbPath := t.TempDir() + "test-db"

	db, err := createDbAndPutCode(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	s := new(leveldb.DBStats)
	err = db.stats(s)
	if err != nil {
		t.Fatalf("cannot get db stats; %v", err)
	}

	// 54 is the base write when creating levelDB
	if s.IOWrite <= 54 {
		t.Fatal("db file should have something inside")
	}

}

func TestCodeDB_HasCode(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutCode(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	has, err := db.HasCode(hash.Keccak256Hash(testCode))
	if err != nil {
		t.Fatalf("get code returned error; %v", err)
	}

	if !has {
		t.Fatal("code is not within db")
	}
}

func TestCodeDB_GetCode(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutCode(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	code, err := db.GetCode(hash.Keccak256Hash(testCode))
	if err != nil {
		t.Fatalf("get code returned error; %v", err)
	}

	if !bytes.Equal(code, testCode) {
		t.Fatal("code returned by the db is different")
	}
}

func TestCodeDB_DeleteCode(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutCode(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	hash := hash.Keccak256Hash(testCode)

	err = db.DeleteCode(hash)
	if err != nil {
		t.Fatalf("delete code returned error; %v", err)
	}

	code, err := db.GetCode(hash)
	if err == nil {
		t.Fatal("get code must fail")
	}

	if got, want := err, leveldb.ErrNotFound; !errors.Is(got, want) {
		t.Fatalf("unexpected err, got: %v, want: %v", got, want)
	}

	if code != nil {
		t.Fatal("code was not deleted")
	}
}

func createDbAndPutCode(dbPath string) (*codeDB, error) {
	db, err := newCodeDB(dbPath, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot open db; %v", err)
	}

	err = db.PutCode(testCode)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestCodeDB_HashCodeSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := &codeDB{
		baseDb, nil, nil,
	}

	input := types.BytesToHash([]byte{1})
	baseDb.EXPECT().Has(CodeDBKey(input), nil).Return(true, nil)
	has, err := db.HasCode(input)
	assert.Nil(t, err)
	assert.Equal(t, true, has)
}

func TestCodeDB_HashCodeEmptyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := &codeDB{
		baseDb, nil, nil,
	}

	input := types.BytesToHash([]byte{})
	has, err := db.HasCode(input)
	assert.Equal(t, ErrorEmptyHash, err)
	assert.Equal(t, false, has)
}

func TestCodeDB_HashCodeReadFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := &codeDB{
		baseDb, nil, nil,
	}

	input := types.BytesToHash([]byte{1})
	baseDb.EXPECT().Has(CodeDBKey(input), nil).Return(false, leveldb.ErrNotFound)
	has, err := db.HasCode(input)
	assert.Equal(t, leveldb.ErrNotFound, err)
	assert.Equal(t, false, has)
}

func TestCodeDB_GetCodeSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := &codeDB{
		baseDb, nil, nil,
	}

	input := types.BytesToHash([]byte{1})
	expectedCode := []byte{1, 2, 3}
	baseDb.EXPECT().Get(CodeDBKey(input), nil).Return(expectedCode, nil)
	code, err := db.GetCode(input)
	assert.Nil(t, err)
	assert.Equal(t, expectedCode, code)
}

func TestCodeDB_GetCodeEmptyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := &codeDB{
		baseDb, nil, nil,
	}

	input := types.BytesToHash([]byte{})
	code, err := db.GetCode(input)
	assert.Equal(t, ErrorEmptyHash, err)
	assert.Nil(t, code)
}

func TestCodeDB_GetCodeReadFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := &codeDB{
		baseDb, nil, nil,
	}

	input := types.BytesToHash([]byte{1})
	baseDb.EXPECT().Get(CodeDBKey(input), nil).Return(nil, leveldb.ErrNotFound)
	code, err := db.GetCode(input)
	assert.NotNil(t, err)
	assert.Nil(t, code)
}

func TestCodeDB_PutCodeSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := &codeDB{
		baseDb, nil, nil,
	}

	inputCode := []byte{1, 2, 3}
	inputHash := hash.Keccak256Hash(inputCode)
	baseDb.EXPECT().Put(CodeDBKey(inputHash), inputCode, nil).Return(nil)
	err := db.PutCode(inputCode)
	assert.Nil(t, err)
}

func TestCodeDB_PutCodeWriteFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := &codeDB{
		baseDb, nil, nil,
	}

	inputCode := []byte{1, 2, 3}
	inputHash := hash.Keccak256Hash(inputCode)
	baseDb.EXPECT().Put(CodeDBKey(inputHash), inputCode, nil).Return(leveldb.ErrReadOnly)
	err := db.PutCode(inputCode)
	assert.NotNil(t, err)
}

func TestCodeDB_DeleteCodeSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := &codeDB{
		baseDb, nil, nil,
	}

	input := types.BytesToHash([]byte{1})
	baseDb.EXPECT().Delete(CodeDBKey(input), nil).Return(nil)
	err := db.DeleteCode(input)
	assert.Nil(t, err)
}

func TestCodeDB_DeleteCodeEmptyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := &codeDB{
		baseDb, nil, nil,
	}

	input := types.BytesToHash([]byte{})
	err := db.DeleteCode(input)
	assert.Equal(t, ErrorEmptyHash, err)
}

func TestCodeDB_DeleteCodeWriteFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := &codeDB{
		baseDb, nil, nil,
	}

	input := types.BytesToHash([]byte{1})
	baseDb.EXPECT().Delete(CodeDBKey(input), nil).Return(leveldb.ErrReadOnly)
	err := db.DeleteCode(input)
	assert.NotNil(t, err)
}

func TestCodeDBKey(t *testing.T) {
	inputHash := hash.Keccak256Hash([]byte{1, 2, 3})
	codeDbKey := CodeDBKey(inputHash)

	assert.Equal(t, uint8(0x31), codeDbKey[0])
	assert.Equal(t, inputHash.Bytes(), codeDbKey[2:])
}

func TestDecodeCodeDBKey(t *testing.T) {
	inputHash := hash.Keccak256Hash([]byte{1, 2, 3})
	codeDbKey := CodeDBKey(inputHash)
	decodedHash, err := DecodeCodeDBKey(codeDbKey)
	assert.Nil(t, err)
	assert.Equal(t, inputHash, decodedHash)
}

func TestDecodeCodeDBKey_InvalidLength(t *testing.T) {
	output, err := DecodeCodeDBKey([]byte("invalid_length_key"))
	assert.NotNil(t, err)
	assert.NotNil(t, output)
}

func TestDecodeCodeDBKey_InvalidPrefix(t *testing.T) {
	output, err := DecodeCodeDBKey([]byte("00" + string(make([]byte, 32))))
	assert.NotNil(t, err)
	assert.NotNil(t, output)
}
