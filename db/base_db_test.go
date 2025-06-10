package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/testutil"
	"github.com/syndtr/goleveldb/leveldb/util"
	"go.uber.org/mock/gomock"
)

func TestBaseDB_BinarySearchForLastPrefixKeySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockDbAdapter(ctrl)
	db := &baseDB{backend: mockBackend}

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

func TestBaseDB_BinarySearchForLastPrefixKeyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockDbAdapter(ctrl)
	db := &baseDB{backend: mockBackend}

	// Case 1: undefined behavior
	kv := &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("value"))
	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv)).Times(9)

	result, err := db.binarySearchForLastPrefixKey([]byte{1})
	assert.NotNil(t, err)
	assert.Equal(t, byte(0), result)
}

func TestBaseDB_HasKeyValuesFor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockDbAdapter(ctrl)
	db := &baseDB{backend: mockBackend}

	// Case 1: Success - found at max value
	kv := &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("value"))
	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv))

	result := db.hasKeyValuesFor([]byte{1}, []byte{1})

	assert.True(t, result)
}

func TestBaseDB_BasicOperations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockDbAdapter(ctrl)
	db := &baseDB{backend: mockBackend}

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

func TestBaseDB_newIterator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockDbAdapter(ctrl)
	mockIter := iterator.NewArrayIterator(&testutil.KeyValue{})
	db := &baseDB{backend: mockBackend}

	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(mockIter)
	iter := db.newIterator(util.BytesPrefix([]byte{1}))
	assert.Equal(t, mockIter, iter)
}

func TestBaseDB_NewIterator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockDbAdapter(ctrl)
	mockIter := iterator.NewArrayIterator(&testutil.KeyValue{})
	db := &baseDB{backend: mockBackend}

	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(mockIter)
	iter := db.NewIterator([]byte("prefix"), []byte("start"))
	assert.Equal(t, mockIter, iter)
}

func TestBaseDB_Compact(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockDbAdapter(ctrl)
	db := &baseDB{backend: mockBackend}

	mockBackend.EXPECT().CompactRange(gomock.Any()).Return(nil)
	err := db.Compact([]byte("start"), []byte("limit"))
	assert.Nil(t, err)
}

func TestBaseDB_Stat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockDbAdapter(ctrl)
	db := &baseDB{backend: mockBackend}

	mockBackend.EXPECT().GetProperty("property").Return("value", nil)
	stat, err := db.Stat("property")
	assert.Nil(t, err)
	assert.Equal(t, "value", stat)
}
