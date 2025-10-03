package db

import (
	"encoding/binary"
	"errors"
	"testing"

	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/testutil"
	"github.com/syndtr/goleveldb/leveldb/util"
	"go.uber.org/mock/gomock"
)

func newTestDestroyedAccountDB(t *testing.T, db DbAdapter, schema SubstateEncodingSchema) *destroyedAccountDB {
	encoding, err := newDestroyedAccountEncoding(schema)
	require.NoError(t, err)
	return &destroyedAccountDB{
		db,
		nil, nil,
		*encoding,
	}
}

func TestDestroyedAccountDB_SetDestroyedAccountsSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	block := uint64(1)
	tx := 2
	destroyed := []types.Address{{1}, {2}}
	resurrected := []types.Address{{3}}

	baseDb.EXPECT().Put(EncodeDestroyedAccountKey(block, tx), gomock.Any(), nil).Return(nil)
	err := db.SetDestroyedAccounts(block, tx, destroyed, resurrected)
	assert.Nil(t, err)
}

func TestDestroyedAccountDB_SetDestroyedAccountsFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	block := uint64(1)
	tx := 2
	destroyed := []types.Address{{1}, {2}}
	resurrected := []types.Address{{3}}

	mockErr := errors.New("mock error")
	baseDb.EXPECT().Put(EncodeDestroyedAccountKey(block, tx), gomock.Any(), nil).Return(mockErr)
	err := db.SetDestroyedAccounts(block, tx, destroyed, resurrected)
	assert.Equal(t, mockErr, err)
}

func TestDestroyedAccountDB_GetDestroyedAccountsSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCases := []struct {
		name     string
		schema   SubstateEncodingSchema
		encodeFn func(SuicidedAccountLists) ([]byte, error)
	}{
		{
			name:     "RLP Encoding",
			schema:   RLPEncodingSchema,
			encodeFn: encodeSuicidedAccountListRLP,
		},
		{
			name:     "PB Encoding",
			schema:   ProtobufEncodingSchema,
			encodeFn: encodeSuicidedAccountListPB,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseDb := NewMockDbAdapter(ctrl)
			db := newTestDestroyedAccountDB(t, baseDb, tc.schema)

			block := uint64(1)
			tx := 2
			expectedDestroyed := []types.Address{{1}, {2}}
			expectedResurrected := []types.Address{{3}}

			list := SuicidedAccountLists{DestroyedAccounts: expectedDestroyed, ResurrectedAccounts: expectedResurrected}
			value, _ := tc.encodeFn(list)

			baseDb.EXPECT().Get(EncodeDestroyedAccountKey(block, tx), nil).Return(value, nil)
			destroyed, resurrected, err := db.GetDestroyedAccounts(block, tx)
			assert.Nil(t, err)
			assert.Equal(t, expectedDestroyed, destroyed)
			assert.Equal(t, expectedResurrected, resurrected)
		})
	}
}

func TestDestroyedAccountDB_GetDestroyedAccountsFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	block := uint64(1)
	tx := 2
	mockErr := errors.New("mock error")

	baseDb.EXPECT().Get(EncodeDestroyedAccountKey(block, tx), nil).Return(nil, nil)
	destroyed, resurrected, err := db.GetDestroyedAccounts(block, tx)
	assert.Nil(t, err)
	assert.Nil(t, destroyed)
	assert.Nil(t, resurrected)

	baseDb.EXPECT().Get(EncodeDestroyedAccountKey(block, tx), nil).Return([]byte{}, mockErr)
	destroyed, resurrected, err = db.GetDestroyedAccounts(block, tx)
	assert.Equal(t, mockErr, err)
	assert.Nil(t, destroyed)
	assert.Nil(t, resurrected)
}

func TestDestroyedAccountDB_GetAccountsDestroyedInRangeSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCases := []struct {
		name     string
		schema   SubstateEncodingSchema
		encodeFn func(SuicidedAccountLists) ([]byte, error)
	}{
		{
			name:     "RLP Encoding",
			schema:   RLPEncodingSchema,
			encodeFn: encodeSuicidedAccountListRLP,
		},
		{
			name:     "PB Encoding",
			schema:   ProtobufEncodingSchema,
			encodeFn: encodeSuicidedAccountListPB,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseDb := NewMockDbAdapter(ctrl)
			kv := &testutil.KeyValue{}
			key1 := EncodeDestroyedAccountKey(5, 0)
			list1 := SuicidedAccountLists{
				DestroyedAccounts:   []types.Address{{1}, {2}},
				ResurrectedAccounts: []types.Address{},
			}
			value1, _ := tc.encodeFn(list1)
			key2 := EncodeDestroyedAccountKey(7, 0)
			list2 := SuicidedAccountLists{
				DestroyedAccounts:   []types.Address{{3}},
				ResurrectedAccounts: []types.Address{{1}},
			}
			value2, _ := tc.encodeFn(list2)
			key3 := EncodeDestroyedAccountKey(99, 0)
			list3 := SuicidedAccountLists{
				DestroyedAccounts:   []types.Address{{3}},
				ResurrectedAccounts: []types.Address{{1}},
			}
			value3, _ := tc.encodeFn(list3)
			kv.PutU(key1, value1)
			kv.PutU(key2, value2)
			kv.PutU(key3, value3)

			iter := iterator.NewArrayIterator(kv)
			db := newTestDestroyedAccountDB(t, baseDb, tc.schema)

			from := uint64(1)
			to := uint64(10)

			startingBlockBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(startingBlockBytes, from)
			r := util.BytesPrefix([]byte(DestroyedAccountPrefix))
			r.Start = append(r.Start, startingBlockBytes...)
			baseDb.EXPECT().NewIterator(r, nil).Return(iter)

			accounts, err := db.GetAccountsDestroyedInRange(from, to)
			assert.NoError(t, err)
			assert.ElementsMatch(t, []types.Address{{2}, {3}}, accounts)
		})
	}
}

func TestDestroyedAccountDB_GetAccountsDestroyedInRangeDecodeKeyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	kv := &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("invalid_key"))
	iter := iterator.NewArrayIterator(kv)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	from := uint64(1)
	to := uint64(10)

	startingBlockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(startingBlockBytes, from)
	r := util.BytesPrefix([]byte(DestroyedAccountPrefix))
	r.Start = append(r.Start, startingBlockBytes...)
	baseDb.EXPECT().NewIterator(r, nil).Return(iter)

	accounts, err := db.GetAccountsDestroyedInRange(from, to)
	assert.NotNil(t, err)
	assert.Nil(t, accounts)
}

func TestDestroyedAccountDB_GetAccountsDestroyedInRangeDecodeValueFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	kv := &testutil.KeyValue{}
	key := EncodeDestroyedAccountKey(5, 0)
	kv.PutU(key, []byte("invalid_key"))
	iter := iterator.NewArrayIterator(kv)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	from := uint64(1)
	to := uint64(10)

	startingBlockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(startingBlockBytes, from)

	r := util.BytesPrefix([]byte(DestroyedAccountPrefix))
	r.Start = append(r.Start, startingBlockBytes...)
	baseDb.EXPECT().NewIterator(r, nil).Return(iter)

	accounts, err := db.GetAccountsDestroyedInRange(from, to)
	assert.NotNil(t, err)
	assert.Nil(t, accounts)
}

func TestDestroyedAccountDB_GetFirstKeySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)

	kv := &testutil.KeyValue{}
	kv.PutU(EncodeDestroyedAccountKey(5, 0), []byte("value0"))
	kv.PutU(EncodeDestroyedAccountKey(6, 1), []byte("value1"))
	iter := iterator.NewArrayIterator(kv)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	r := util.BytesPrefix([]byte(DestroyedAccountPrefix))
	baseDb.EXPECT().NewIterator(r, nil).Return(iter)

	block, err := db.GetFirstKey()
	assert.Nil(t, err)
	assert.Equal(t, uint64(5), block)
}

func TestDestroyedAccountDB_GetFirstKeyDecodeFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)

	// case decode key fail
	kv := &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("value"))
	iter := iterator.NewArrayIterator(kv)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)
	r := util.BytesPrefix([]byte(DestroyedAccountPrefix))
	baseDb.EXPECT().NewIterator(r, nil).Return(iter)
	block, err := db.GetFirstKey()
	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), block)

	// case empty iterator
	kv = &testutil.KeyValue{}
	iter = iterator.NewArrayIterator(kv)
	baseDb.EXPECT().NewIterator(r, nil).Return(iter)
	block, err = db.GetFirstKey()
	assert.Equal(t, leveldb.ErrNotFound, err)
	assert.Equal(t, uint64(0), block)
}

func TestDestroyedAccountDB_GetFirstKeyNoUpdateSetFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	kv := &testutil.KeyValue{}
	iter := iterator.NewArrayIterator(kv)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	r := util.BytesPrefix([]byte(DestroyedAccountPrefix))
	baseDb.EXPECT().NewIterator(r, nil).Return(iter)

	block, err := db.GetFirstKey()
	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), block)
}

func TestDestroyedAccountDB_GetLastKeySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	kv := &testutil.KeyValue{}
	kv.PutU(EncodeDestroyedAccountKey(5, 0), []byte("value0"))
	kv.PutU(EncodeDestroyedAccountKey(10, 1), []byte("value1"))
	iter := iterator.NewArrayIterator(kv)

	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	r := util.BytesPrefix([]byte(DestroyedAccountPrefix))
	baseDb.EXPECT().NewIterator(r, nil).Return(iter)

	block, err := db.GetLastKey()
	assert.Nil(t, err)
	assert.Equal(t, uint64(10), block)
}

func TestDestroyedAccountDB_GetLastKeyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)

	// case decode key fail
	kv := &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("invalid_key"))
	iter := iterator.NewArrayIterator(kv)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	r := util.BytesPrefix([]byte(DestroyedAccountPrefix))
	baseDb.EXPECT().NewIterator(r, nil).Return(iter)

	block, err := db.GetLastKey()
	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), block)

	// case empty iterator
	kv = &testutil.KeyValue{}
	iter = iterator.NewArrayIterator(kv)
	baseDb.EXPECT().NewIterator(r, nil).Return(iter)
	block, err = db.GetLastKey()
	assert.Equal(t, leveldb.ErrNotFound, err)
	assert.Equal(t, uint64(0), block)
}

func TestDestroyedAccountDB_BasicOperations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockDbAdapter(ctrl)
	db := newTestDestroyedAccountDB(t, mockBackend, DefaultEncodingSchema)

	// Test stats
	stats := &leveldb.DBStats{}
	mockBackend.EXPECT().Stats(gomock.Any()).Return(nil)
	err := db.stats(stats)
	assert.Nil(t, err)

	// Test GetBackend
	backend := db.GetBackend()
	assert.Equal(t, mockBackend, backend)

	// Test Put
	mockBackend.EXPECT().Put([]byte("key"), []byte("value"), gomock.Any()).Return(nil)
	err = db.Put([]byte("key"), []byte("value"))
	assert.Nil(t, err)

	// Test Get
	mockBackend.EXPECT().Get([]byte("key"), gomock.Any()).Return([]byte("value"), nil)
	value, err := db.Get([]byte("key"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("value"), value)

	// Test Has
	mockBackend.EXPECT().Has([]byte("key"), gomock.Any()).Return(true, nil)
	exists, err := db.Has([]byte("key"))
	assert.Nil(t, err)
	assert.True(t, exists)

	// Test Delete
	mockBackend.EXPECT().Delete([]byte("key"), gomock.Any()).Return(nil)
	err = db.Delete([]byte("key"))
	assert.Nil(t, err)

	// Test Close
	mockBackend.EXPECT().Close().Return(nil)
	err = db.Close()
	assert.Nil(t, err)

	// Test NewBatch
	b := db.NewBatch()
	assert.NotNil(t, b)
}

func TestDestroyedAccountDB_newIterator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockDbAdapter(ctrl)
	mockIter := iterator.NewArrayIterator(&testutil.KeyValue{})
	db := newTestDestroyedAccountDB(t, mockBackend, DefaultEncodingSchema)

	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(mockIter)
	iter := db.newIterator(util.BytesPrefix([]byte{1}))
	assert.Equal(t, mockIter, iter)
}

func TestDestroyedAccountDB_Compact(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockDbAdapter(ctrl)
	db := newTestDestroyedAccountDB(t, mockBackend, DefaultEncodingSchema)

	mockBackend.EXPECT().CompactRange(gomock.Any()).Return(nil)
	err := db.Compact([]byte("start"), []byte("limit"))
	assert.Nil(t, err)
}

func TestDestroyedAccountDB_Stat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockDbAdapter(ctrl)
	db := newTestDestroyedAccountDB(t, mockBackend, DefaultEncodingSchema)

	mockBackend.EXPECT().GetProperty("property").Return("value", nil)
	stat, err := db.Stat("property")
	assert.Nil(t, err)
	assert.Equal(t, "value", stat)
}

func TestDestroyedAccountDB_BinarySearchForLastPrefixKeySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockDbAdapter(ctrl)
	db := newTestDestroyedAccountDB(t, mockBackend, DefaultEncodingSchema)

	// Case 1: found at max value
	kv := &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("value"))
	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv)).Times(8)
	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv))
	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv))

	result, err := db.binarySearchForLastPrefixKey([]byte{1})
	assert.Nil(t, err)
	assert.Equal(t, byte(0x80), result)

	// Case 2: found at min value
	kv = &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("value"))
	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv)).Times(8)
	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv)).Times(2)

	result, err = db.binarySearchForLastPrefixKey([]byte{1})
	assert.Nil(t, err)
	assert.Equal(t, byte(0x7f), result)
}

func TestDestroyedAccountDB_BinarySearchForLastPrefixKeyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockDbAdapter(ctrl)
	db := newTestDestroyedAccountDB(t, mockBackend, DefaultEncodingSchema)

	// Case 1: undefined behavior
	kv := &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("value"))
	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv)).Times(9)

	result, err := db.binarySearchForLastPrefixKey([]byte{1})
	assert.NotNil(t, err)
	assert.Equal(t, byte(0), result)
}

func TestDestroyedAccountDB_HasKeyValuesFor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockDbAdapter(ctrl)
	db := newTestDestroyedAccountDB(t, mockBackend, DefaultEncodingSchema)

	// Case 1: Success - found at max value
	kv := &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("value"))
	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv))

	result := db.hasKeyValuesFor([]byte{1}, []byte{1})

	assert.True(t, result)
}

func TestDestroyedAccountDB_encodeDestroyedAccountKey(t *testing.T) {
	block := uint64(123456789)
	tx := 42
	key := EncodeDestroyedAccountKey(block, tx)

	expected := append([]byte(DestroyedAccountPrefix), make([]byte, 12)...)
	binary.BigEndian.PutUint64(expected[len(DestroyedAccountPrefix):], block)
	binary.BigEndian.PutUint32(expected[len(DestroyedAccountPrefix)+8:], uint32(tx))

	assert.Equal(t, expected, key)
}

func TestDestroyedAccountDB_DecodeDestroyedAccountKeySuccess(t *testing.T) {
	block := uint64(123456789)
	tx := 42
	key := EncodeDestroyedAccountKey(block, tx)

	decodedBlock, decodedTx, err := DecodeDestroyedAccountKey(key)
	assert.Nil(t, err)
	assert.Equal(t, block, decodedBlock)
	assert.Equal(t, tx, decodedTx)
}

func TestDestroyedAccountDB_DecodeDestroyedAccountKeyFail(t *testing.T) {
	// Invalid length
	_, _, err := DecodeDestroyedAccountKey([]byte("invalid"))
	assert.NotNil(t, err)

	// Invalid prefix
	invalidPrefix := make([]byte, len(DestroyedAccountPrefix)+12)
	copy(invalidPrefix, "xx")
	_, _, err = DecodeDestroyedAccountKey(invalidPrefix)
	assert.NotNil(t, err)
}
