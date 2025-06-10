package db

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/mock/gomock"
)

func TestReplayer_Put(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockKeyValueWriter(ctrl)
	r := &replayer{writer: mockWriter}

	// Test successful put
	mockWriter.EXPECT().Put([]byte("key"), []byte("value")).Return(nil)
	r.Put([]byte("key"), []byte("value"))
	assert.Nil(t, r.failure)

	// Test failed put
	expectedErr := errors.New("put error")
	mockWriter.EXPECT().Put([]byte("key2"), []byte("value2")).Return(expectedErr)
	r.Put([]byte("key2"), []byte("value2"))
	assert.Equal(t, expectedErr, r.failure)

	// Test that no operation is performed after failure
	r.Put([]byte("key3"), []byte("value3"))
	// No expectations set for mockWriter - should not be called
}

func TestReplayer_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockKeyValueWriter(ctrl)
	r := &replayer{writer: mockWriter}

	// Test successful delete
	mockWriter.EXPECT().Delete([]byte("key")).Return(nil)
	r.Delete([]byte("key"))
	assert.Nil(t, r.failure)

	// Test failed delete
	expectedErr := errors.New("delete error")
	mockWriter.EXPECT().Delete([]byte("key2")).Return(expectedErr)
	r.Delete([]byte("key2"))
	assert.Equal(t, expectedErr, r.failure)

	// Test that no operation is performed after failure
	r.Delete([]byte("key3"))
	// No expectations set for mockWriter - should not be called
}

func TestBatch_Put(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := NewMockDbAdapter(ctrl)
	b := &batch{
		db: mockAdapter,
		b:  new(leveldb.Batch),
	}

	key := []byte("key")
	value := []byte("value")
	err := b.Put(key, value)

	assert.Nil(t, err)
	assert.Equal(t, len(value), b.size)
}

func TestBatch_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := NewMockDbAdapter(ctrl)
	b := &batch{
		db: mockAdapter,
		b:  new(leveldb.Batch),
	}

	key := []byte("key")
	err := b.Delete(key)

	assert.Nil(t, err)
	assert.Equal(t, len(key), b.size)
}

func TestBatch_ValueSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := NewMockDbAdapter(ctrl)
	b := &batch{
		db:   mockAdapter,
		b:    new(leveldb.Batch),
		size: 100,
	}

	size := b.ValueSize()
	assert.Equal(t, 100, size)
}

func TestBatch_Write(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := NewMockDbAdapter(ctrl)
	b := &batch{
		db: mockAdapter,
		b:  new(leveldb.Batch),
	}

	// Test successful write
	mockAdapter.EXPECT().Write(b.b, nil).Return(nil)
	err := b.Write()
	assert.Nil(t, err)

	// Test failed write
	expectedErr := errors.New("write error")
	mockAdapter.EXPECT().Write(b.b, nil).Return(expectedErr)
	err = b.Write()
	assert.Equal(t, expectedErr, err)
}

func TestBatch_Reset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := NewMockDbAdapter(ctrl)
	b := &batch{
		db:   mockAdapter,
		b:    new(leveldb.Batch),
		size: 100,
	}

	b.Reset()
	assert.Equal(t, 0, b.size)
}

func TestBatch_Replay(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := NewMockDbAdapter(ctrl)
	mockWriter := NewMockKeyValueWriter(ctrl)

	b := &batch{
		db: mockAdapter,
		b:  new(leveldb.Batch),
	}

	// Add some operations to the batch
	err := b.Put([]byte("key1"), []byte("value1"))
	if err != nil {
		t.Fatal(err)
	}

	// Test successful replay
	mockWriter.EXPECT().Put([]byte("key1"), []byte("value1")).Return(nil)

	err = b.Replay(mockWriter)
	assert.Nil(t, err)
}

func TestNewBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := NewMockDbAdapter(ctrl)
	b := newBatch(mockAdapter)

	assert.NotNil(t, b)
	batchImpl, ok := b.(*batch)
	assert.True(t, ok)
	assert.Equal(t, mockAdapter, batchImpl.db)
	assert.NotNil(t, batchImpl.b)
	assert.Equal(t, 0, batchImpl.size)
}
