package db

import (
	"errors"
	"fmt"
	"testing"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/testutil"
	"go.uber.org/mock/gomock"
)

var testUpdateSet = &updateset.UpdateSet{
	WorldState: substate.WorldState{
		types.Address{1}: &substate.Account{
			Nonce:   1,
			Balance: new(uint256.Int).SetUint64(1),
		},
		types.Address{2}: &substate.Account{
			Nonce:   2,
			Balance: new(uint256.Int).SetUint64(2),
		},
	},
	Block: 1,
}

var testDeletedAccounts = []types.Address{{3}, {4}}

func newTestUpdateDB(t *testing.T, db *MockCodeDB, schema SubstateEncodingSchema) *updateDB {
	encoding, err := newUpdateSetEncoding(schema)
	require.NoError(t, err)
	return &updateDB{
		db,
		*encoding,
	}
}

func TestUpdateDB_PutUpdateSet(t *testing.T) {
	testCases := []struct {
		name   string
		schema SubstateEncodingSchema
	}{
		{"RLP", RLPEncodingSchema},
		{"PB", ProtobufEncodingSchema},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dbPath := t.TempDir() + "test-db"
			db, err := createDbAndPutUpdateSet(dbPath, tc.schema)
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
		})
	}
}

func TestUpdateDB_HasUpdateSet(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutUpdateSet(dbPath, DefaultEncodingSchema)
	if err != nil {
		t.Fatal(err)
	}

	has, err := db.HasUpdateSet(testUpdateSet.Block)
	if err != nil {
		t.Fatalf("has update-set returned error; %v", err)
	}

	if !has {
		t.Fatal("update-set is not within db")
	}
}

func TestUpdateDB_GetUpdateSet(t *testing.T) {
	testCases := []struct {
		name   string
		schema SubstateEncodingSchema
	}{
		{"RLP", RLPEncodingSchema},
		{"PB", ProtobufEncodingSchema},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dbPath := t.TempDir() + "test-db"
			db, err := createDbAndPutUpdateSet(dbPath, tc.schema)
			if err != nil {
				t.Fatal(err)
			}

			us, err := db.GetUpdateSet(testUpdateSet.Block)
			if err != nil {
				t.Fatalf("get update-set returned error; %v", err)
			}

			if us == nil {
				t.Fatal("update-set is nil")
			}

			if !us.Equal(testUpdateSet) {
				t.Fatal("substates are different")
			}
		})
	}
}

func TestUpdateDB_DeleteUpdateSet(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutUpdateSet(dbPath, DefaultEncodingSchema)
	if err != nil {
		t.Fatal(err)
	}

	err = db.DeleteUpdateSet(testUpdateSet.Block)
	if err != nil {
		t.Fatalf("delete update-set returned error; %v", err)
	}

	us, err := db.GetUpdateSet(testUpdateSet.Block)
	if err == nil {
		t.Fatal("get update-set must fail")
	}

	if got, want := err, leveldb.ErrNotFound; !errors.Is(got, want) {
		t.Fatalf("unexpected err, got: %v, want: %v", got, want)
	}

	if us != nil {
		t.Fatal("update-set was not deleted")
	}
}

func TestUpdateDB_GetFirstKey(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutUpdateSet(dbPath, DefaultEncodingSchema)
	if err != nil {
		t.Fatal(err)
	}

	got, err := db.GetFirstKey()
	if err != nil {
		t.Fatalf("cannot get first key; %v", err)
	}

	var want = testUpdateSet.Block

	if want != got {
		t.Fatalf("incorrect first key\nwant: %v\ngot: %v", want, got)
	}
}

func TestUpdateDB_GetLastKey(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutUpdateSet(dbPath, DefaultEncodingSchema)
	if err != nil {
		t.Fatal(err)
	}

	got, err := db.GetLastKey()
	if err != nil {
		t.Fatalf("cannot get last key; %v", err)
	}

	var want = testUpdateSet.Block

	if want != got {
		t.Fatalf("incorrect last key\nwant: %v\ngot: %v", want, got)
	}
}

func createDbAndPutUpdateSet(dbPath string, encoding SubstateEncodingSchema) (*updateDB, error) {
	db, err := newUpdateDB(dbPath, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot open db; %v", err)
	}
	err = db.SetSubstateEncoding(encoding)
	if err != nil {
		return nil, fmt.Errorf("cannot set encoding; %v", err)
	}
	err = db.PutUpdateSet(testUpdateSet, testDeletedAccounts)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestUpdateDB_GetFirstKeySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)
	kv := &testutil.KeyValue{}
	kv.PutU(UpdateDBKey(1), []byte{42})
	kv.PutU(UpdateDBKey(2), []byte{43})
	mockIter := iterator.NewArrayIterator(kv)

	mockDB.EXPECT().newIterator(gomock.Any()).Return(mockIter)

	db := newTestUpdateDB(t, mockDB, ProtobufEncodingSchema)

	result, err := db.GetFirstKey()

	assert.Nil(t, err)
	assert.Equal(t, uint64(1), result)
}

func TestUpdateDB_GetFirstKeyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)

	// case 1 not found
	kv := &testutil.KeyValue{}
	mockIter := iterator.NewArrayIterator(kv)
	db := newTestUpdateDB(t, mockDB, ProtobufEncodingSchema)

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
func TestUpdateDB_GetLastKeySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)
	kv := &testutil.KeyValue{}
	kv.PutU(UpdateDBKey(1), []byte{30})
	kv.PutU(UpdateDBKey(5), []byte{42})
	mockIter := iterator.NewArrayIterator(kv)

	mockDB.EXPECT().newIterator(gomock.Any()).Return(mockIter)

	db := newTestUpdateDB(t, mockDB, ProtobufEncodingSchema)

	result, err := db.GetLastKey()

	assert.Nil(t, err)
	assert.Equal(t, uint64(5), result)
}

func TestUpdateDB_GetLastKeyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)

	// case 1: no updateset found
	kv := &testutil.KeyValue{}
	mockIter := iterator.NewArrayIterator(kv)
	db := newTestUpdateDB(t, mockDB, ProtobufEncodingSchema)

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

func TestUpdateDB_HasUpdateSetSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)
	blockNum := uint64(10)
	key := UpdateDBKey(blockNum)

	mockDB.EXPECT().Has(key).Return(true, nil)

	db := newTestUpdateDB(t, mockDB, ProtobufEncodingSchema)
	result, err := db.HasUpdateSet(blockNum)

	assert.Nil(t, err)
	assert.True(t, result)
}

func TestUpdateDB_HasUpdateSetFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)
	blockNum := uint64(10)
	key := UpdateDBKey(blockNum)
	expectedErr := errors.New("database error")

	mockDB.EXPECT().Has(key).Return(false, expectedErr)

	db := newTestUpdateDB(t, mockDB, ProtobufEncodingSchema)
	result, err := db.HasUpdateSet(blockNum)

	assert.Equal(t, expectedErr, err)
	assert.False(t, result)
}

func TestUpdateDB_GetUpdateSetSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)
	blockNum := uint64(10)
	key := UpdateDBKey(blockNum)

	updateSet := &updateset.UpdateSet{
		WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
		Block:           0,
		DeletedAccounts: []types.Address{},
	}
	encodedData, _ := encodeUpdateSetPB(*updateSet, []types.Address{{}})

	db := newTestUpdateDB(t, mockDB, ProtobufEncodingSchema)

	// case 1: Get success
	mockDB.EXPECT().Get(key).Return(encodedData, nil)
	mockDB.EXPECT().GetCode(gomock.Any()).Return(nil, nil).AnyTimes()

	result, err := db.GetUpdateSet(blockNum)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, blockNum, result.Block)

	// case 2: Get success nil
	mockDB.EXPECT().Get(key).Return(nil, nil)

	result, err = db.GetUpdateSet(blockNum)
	assert.Nil(t, err)
	assert.Nil(t, result)
}

func TestUpdateDB_GetUpdateSetFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)
	blockNum := uint64(10)
	key := UpdateDBKey(blockNum)

	// Case 1: Get error
	expectedErr := errors.New("database error")
	mockDB.EXPECT().Get(key).Return(nil, expectedErr)

	db := newTestUpdateDB(t, mockDB, ProtobufEncodingSchema)
	result, err := db.GetUpdateSet(blockNum)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), expectedErr.Error())

	// Case 2: decode error
	mockDB.EXPECT().Get(key).Return([]byte{1, 2, 3}, nil) // Invalid RLP data

	result, err = db.GetUpdateSet(blockNum)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cannot decode update-set")
}
func TestUpdateDB_PutUpdateSetSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)

	updateSet := &updateset.UpdateSet{
		WorldState: substate.WorldState{
			types.Address{1}: &substate.Account{
				Nonce:   1,
				Balance: new(uint256.Int).SetUint64(1),
				Code:    []byte{0x01, 0x02},
			},
		},
		Block: 10,
	}

	deletedAccounts := []types.Address{{2}}
	key := UpdateDBKey(updateSet.Block)

	// Expectations
	mockDB.EXPECT().PutCode(gomock.Any()).Return(nil)
	mockDB.EXPECT().Put(key, gomock.Any()).Return(nil)

	db := newTestUpdateDB(t, mockDB, ProtobufEncodingSchema)
	err := db.PutUpdateSet(updateSet, deletedAccounts)

	assert.Nil(t, err)
}

func TestUpdateDB_PutUpdateSetFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)

	updateSet := &updateset.UpdateSet{
		WorldState: substate.WorldState{
			types.Address{1}: &substate.Account{
				Nonce:   1,
				Balance: new(uint256.Int).SetUint64(1),
				Code:    []byte{0x01, 0x02},
			},
		},
		Block: 10,
	}

	deletedAccounts := []types.Address{{2}}
	expectedErr := errors.New("code storage error")

	// Case 1: PutCode error
	mockDB.EXPECT().PutCode(gomock.Any()).Return(expectedErr)

	db := newTestUpdateDB(t, mockDB, ProtobufEncodingSchema)
	err := db.PutUpdateSet(updateSet, deletedAccounts)

	assert.Equal(t, expectedErr, err)

	// Case 2: Put error
	mockDB.EXPECT().PutCode(gomock.Any()).Return(nil)
	mockDB.EXPECT().Put(gomock.Any(), gomock.Any()).Return(expectedErr)

	err = db.PutUpdateSet(updateSet, deletedAccounts)

	assert.Equal(t, expectedErr, err)
}
func TestUpdateDB_DeleteUpdateSetSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)
	blockNum := uint64(10)
	key := UpdateDBKey(blockNum)

	mockDB.EXPECT().Delete(key).Return(nil)

	db := newTestUpdateDB(t, mockDB, ProtobufEncodingSchema)
	err := db.DeleteUpdateSet(blockNum)

	assert.Nil(t, err)
}

func TestUpdateDB_DeleteUpdateSetFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)
	blockNum := uint64(10)
	key := UpdateDBKey(blockNum)
	expectedErr := errors.New("delete error")

	mockDB.EXPECT().Delete(key).Return(expectedErr)

	db := newTestUpdateDB(t, mockDB, ProtobufEncodingSchema)
	err := db.DeleteUpdateSet(blockNum)

	assert.Equal(t, expectedErr, err)
}

func TestUpdateDB_NewUpdateSetIterator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)
	db := newTestUpdateDB(t, mockDB, ProtobufEncodingSchema)

	start := uint64(1)
	end := uint64(4)

	// Create a mock iterator that would be returned internally
	kv := &testutil.KeyValue{}
	updateSet := &updateset.UpdateSet{
		WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
		Block:           0,
		DeletedAccounts: []types.Address{},
	}
	encodedData, err := encodeUpdateSetPB(*updateSet, []types.Address{{}})
	if err != nil {
		t.Fatal(err)
	}
	kv.PutU(UpdateDBKey(1), encodedData)
	kv.PutU(UpdateDBKey(2), encodedData)
	kv.PutU(UpdateDBKey(3), encodedData)
	kv.PutU(UpdateDBKey(4), encodedData)
	mockIter := iterator.NewArrayIterator(kv)

	// Set up expectations for the newUpdateSetIterator behavior
	// This is testing that the method correctly initializes the iterator with the right parameters
	mockDB.EXPECT().newIterator(gomock.Any()).Return(mockIter)
	mockDB.EXPECT().GetCode(gomock.Any()).Return([]byte("code"), nil).AnyTimes()

	// Test the iterator
	iter := db.NewUpdateSetIterator(start, end)

	// Verify iterator behavior
	results := make([]*updateset.UpdateSet, 0)
	for iter.Next() {
		results = append(results, iter.Value())
	}
	iter.Release()
	assert.Nil(t, iter.Error())
	assert.Len(t, results, 4)
}

func TestDecodeUpdateSetKey_Success(t *testing.T) {
	blockNum := uint64(12345)
	key := UpdateDBKey(blockNum)

	result, err := DecodeUpdateSetKey(key)

	assert.Nil(t, err)
	assert.Equal(t, blockNum, result)
}

func TestDecodeUpdateSetKey_Fail(t *testing.T) {
	// Case 1: invalid key length
	key := []byte(UpdateDBPrefix + "short")
	result, err := DecodeUpdateSetKey(key)

	assert.Error(t, err)
	assert.Equal(t, uint64(0), result)
	assert.Contains(t, err.Error(), "invalid length")

	// Case 2: invalid prefix
	key = []byte("XX" + string(make([]byte, 8)))
	result, err = DecodeUpdateSetKey(key)

	assert.Error(t, err)
	assert.Equal(t, uint64(0), result)
	assert.Contains(t, err.Error(), "invalid prefix")
}

func TestUpdateDBKey(t *testing.T) {
	blockNum := uint64(12345)
	result := UpdateDBKey(blockNum)

	// Key should be prefix + 8 bytes for block number
	assert.Equal(t, len(UpdateDBPrefix)+8, len(result))
	assert.Equal(t, []byte(UpdateDBPrefix), result[:len(UpdateDBPrefix)])

	// Decode to verify
	decoded, err := DecodeUpdateSetKey(result)
	assert.Nil(t, err)
	assert.Equal(t, blockNum, decoded)
}
