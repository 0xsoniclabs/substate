package db

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/testutil"
)

func TestIterator_Next(t *testing.T) {
	mockIter := iterator.NewArrayIterator(&testutil.KeyValue{})
	iter := newIterator[string](mockIter)

	go func() {
		time.Sleep(5 * time.Millisecond)
		iter.resultCh <- "result1"
		time.Sleep(5 * time.Millisecond)
		iter.resultCh <- ""
	}()
	r1 := iter.Next()
	assert.True(t, r1)
	r2 := iter.Next()
	assert.False(t, r2)
}

func TestIterator_NextWithError(t *testing.T) {
	mockIter := iterator.NewArrayIterator(&testutil.KeyValue{})
	iter := newIterator[string](mockIter)

	iter.err = errors.New("mock error")
	r1 := iter.Next()
	assert.False(t, r1)
}

func TestIterator_Error(t *testing.T) {
	mockIter := iterator.NewArrayIterator(&testutil.KeyValue{})
	iter := newIterator[string](mockIter)

	mockErr := errors.New("mock error")
	iter.err = mockErr
	r1 := iter.Error()
	assert.NotNil(t, r1)
	assert.Equal(t, mockErr.Error(), r1.Error())
}

func TestIterator_Value(t *testing.T) {
	mockIter := iterator.NewArrayIterator(&testutil.KeyValue{})
	iter := newIterator[string](mockIter)

	mockInput := "result1"
	iter.cur = mockInput
	r1 := iter.Value()
	assert.NotNil(t, r1)
	assert.Equal(t, mockInput, r1)
}

func TestIterator_Release(t *testing.T) {
	mockIter := iterator.NewArrayIterator(&testutil.KeyValue{})
	iter := newIterator[string](mockIter)
	iter.Release()
}
