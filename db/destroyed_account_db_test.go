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
	"go.uber.org/mock/gomock"
)

func newTestDestroyedAccountDB(t *testing.T, db BaseDB, schema SubstateEncodingSchema) *destroyedAccountDB {
	encoding, err := newDestroyedAccountEncoding(schema)
	require.NoError(t, err)
	return &destroyedAccountDB{
		db,
		*encoding,
	}
}

func TestDestroyedAccountDB_SetDestroyedAccountsSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockBaseDB(ctrl)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	block := uint64(1)
	tx := 2
	destroyed := []types.Address{{1}, {2}}
	resurrected := []types.Address{{3}}

	baseDb.EXPECT().Put(EncodeDestroyedAccountKey(block, tx), gomock.Any()).Return(nil)
	err := db.SetDestroyedAccounts(block, tx, destroyed, resurrected)
	assert.Nil(t, err)
}

func TestDestroyedAccountDB_SetDestroyedAccountsFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockBaseDB(ctrl)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	block := uint64(1)
	tx := 2
	destroyed := []types.Address{{1}, {2}}
	resurrected := []types.Address{{3}}

	mockErr := errors.New("mock error")
	baseDb.EXPECT().Put(EncodeDestroyedAccountKey(block, tx), gomock.Any()).Return(mockErr)
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
			baseDb := NewMockBaseDB(ctrl)
			db := newTestDestroyedAccountDB(t, baseDb, tc.schema)

			block := uint64(1)
			tx := 2
			expectedDestroyed := []types.Address{{1}, {2}}
			expectedResurrected := []types.Address{{3}}

			list := SuicidedAccountLists{DestroyedAccounts: expectedDestroyed, ResurrectedAccounts: expectedResurrected}
			value, _ := tc.encodeFn(list)

			baseDb.EXPECT().Get(EncodeDestroyedAccountKey(block, tx)).Return(value, nil)
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

	baseDb := NewMockBaseDB(ctrl)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	block := uint64(1)
	tx := 2
	mockErr := errors.New("mock error")

	baseDb.EXPECT().Get(EncodeDestroyedAccountKey(block, tx)).Return(nil, nil)
	destroyed, resurrected, err := db.GetDestroyedAccounts(block, tx)
	assert.Nil(t, err)
	assert.Nil(t, destroyed)
	assert.Nil(t, resurrected)

	baseDb.EXPECT().Get(EncodeDestroyedAccountKey(block, tx)).Return([]byte{}, mockErr)
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
			baseDb := NewMockBaseDB(ctrl)
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

			baseDb.EXPECT().NewIterator([]byte(DestroyedAccountPrefix), startingBlockBytes).Return(iter)

			accounts, err := db.GetAccountsDestroyedInRange(from, to)
			assert.NoError(t, err)
			assert.ElementsMatch(t, []types.Address{{2}, {3}}, accounts)
		})
	}
}

func TestDestroyedAccountDB_GetAccountsDestroyedInRangeDecodeKeyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockBaseDB(ctrl)
	kv := &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("invalid_key"))
	iter := iterator.NewArrayIterator(kv)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	from := uint64(1)
	to := uint64(10)

	startingBlockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(startingBlockBytes, from)

	baseDb.EXPECT().NewIterator([]byte(DestroyedAccountPrefix), startingBlockBytes).Return(iter)

	accounts, err := db.GetAccountsDestroyedInRange(from, to)
	assert.NotNil(t, err)
	assert.Nil(t, accounts)
}

func TestDestroyedAccountDB_GetAccountsDestroyedInRangeDecodeValueFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockBaseDB(ctrl)
	kv := &testutil.KeyValue{}
	key := EncodeDestroyedAccountKey(5, 0)
	kv.PutU(key, []byte("invalid_key"))
	iter := iterator.NewArrayIterator(kv)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	from := uint64(1)
	to := uint64(10)

	startingBlockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(startingBlockBytes, from)

	baseDb.EXPECT().NewIterator([]byte(DestroyedAccountPrefix), startingBlockBytes).Return(iter)

	accounts, err := db.GetAccountsDestroyedInRange(from, to)
	assert.NotNil(t, err)
	assert.Nil(t, accounts)
}

func TestDestroyedAccountDB_GetFirstKeySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockBaseDB(ctrl)

	kv := &testutil.KeyValue{}
	kv.PutU(EncodeDestroyedAccountKey(5, 0), []byte("value0"))
	kv.PutU(EncodeDestroyedAccountKey(6, 1), []byte("value1"))
	iter := iterator.NewArrayIterator(kv)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	baseDb.EXPECT().NewIterator([]byte(DestroyedAccountPrefix), nil).Return(iter)

	block, err := db.GetFirstKey()
	assert.Nil(t, err)
	assert.Equal(t, uint64(5), block)
}

func TestDestroyedAccountDB_GetFirstKeyDecodeFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockBaseDB(ctrl)

	// case decode key fail
	kv := &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("value"))
	iter := iterator.NewArrayIterator(kv)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)
	baseDb.EXPECT().NewIterator([]byte(DestroyedAccountPrefix), nil).Return(iter)
	block, err := db.GetFirstKey()
	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), block)

	// case empty iterator
	kv = &testutil.KeyValue{}
	iter = iterator.NewArrayIterator(kv)
	baseDb.EXPECT().NewIterator([]byte(DestroyedAccountPrefix), nil).Return(iter)
	block, err = db.GetFirstKey()
	assert.Equal(t, leveldb.ErrNotFound, err)
	assert.Equal(t, uint64(0), block)
}

func TestDestroyedAccountDB_GetFirstKeyNoUpdateSetFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockBaseDB(ctrl)
	kv := &testutil.KeyValue{}
	iter := iterator.NewArrayIterator(kv)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	baseDb.EXPECT().NewIterator([]byte(DestroyedAccountPrefix), nil).Return(iter)

	block, err := db.GetFirstKey()
	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), block)
}

func TestDestroyedAccountDB_GetLastKeySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockBaseDB(ctrl)
	kv := &testutil.KeyValue{}
	kv.PutU(EncodeDestroyedAccountKey(5, 0), []byte("value0"))
	kv.PutU(EncodeDestroyedAccountKey(10, 1), []byte("value1"))
	iter := iterator.NewArrayIterator(kv)

	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	baseDb.EXPECT().NewIterator([]byte(DestroyedAccountPrefix), nil).Return(iter)

	block, err := db.GetLastKey()
	assert.Nil(t, err)
	assert.Equal(t, uint64(10), block)
}

func TestDestroyedAccountDB_GetLastKeyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockBaseDB(ctrl)

	// case decode key fail
	kv := &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("invalid_key"))
	iter := iterator.NewArrayIterator(kv)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)
	baseDb.EXPECT().NewIterator([]byte(DestroyedAccountPrefix), nil).Return(iter)

	block, err := db.GetLastKey()
	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), block)

	// case empty iterator
	kv = &testutil.KeyValue{}
	iter = iterator.NewArrayIterator(kv)
	baseDb.EXPECT().NewIterator([]byte(DestroyedAccountPrefix), nil).Return(iter)
	block, err = db.GetLastKey()
	assert.Equal(t, leveldb.ErrNotFound, err)
	assert.Equal(t, uint64(0), block)
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
