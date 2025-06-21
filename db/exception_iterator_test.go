package db

import (
	"encoding/binary"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/testutil"
	"go.uber.org/mock/gomock"
)

func TestExceptionIterator_Next(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutException(path)
	if err != nil {
		return
	}

	iter := db.NewExceptionIterator(0, 10)
	if !iter.Next() {
		t.Fatal("next must return true")
	}

	if iter.Next() {
		t.Fatal("next must return false, all exceptions were extracted")
	}
}

func TestExceptionIterator_Value(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutException(path)
	if err != nil {
		return
	}

	testExceptionIterator_Value(db, t)
}

func testExceptionIterator_Value(db *exceptionDB, t *testing.T) {
	iter := db.NewExceptionIterator(0, 10)

	if !iter.Next() {
		t.Fatal("next must return true")
	}

	ex := iter.Value()

	if ex == nil {
		t.Fatal("iterator returned nil")
	}

	if !ex.Equal(*testException) {
		t.Fatalf("iterator returned different exception.")
	}
}

func TestExceptionIterator_FromBlock(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutException(path)
	if err != nil {
		t.Fatal(err)
	}

	test2 := substate.Exception{Block: testException.Block, Data: testException.Data}
	test2.Block++

	err = db.PutException(&test2)
	if err != nil {
		t.Fatalf("unable to put exception: %v", err)
	}

	iter := db.NewExceptionIterator(int(testException.Block), 10)

	if !iter.Next() {
		t.Fatal("next must return true")
	}

	ss := iter.Value()
	if ss.Block != testException.Block {
		t.Fatal("incorrect block number")
	}

	counter := 1
	for iter.Next() {
		counter++
	}

	if counter != 2 {
		t.Fatal("incorrect number of exceptions")
	}

	iter2 := db.NewExceptionIterator(int(testException.Block), 10)

	if !iter2.Next() {
		t.Fatal("next must return true")
	}

	ss2 := iter2.Value()

	if ss2 == nil {
		t.Fatal("iterator returned nil")
	}
}

func TestExceptionIterator_Release(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutException(path)
	if err != nil {
		return
	}

	iter := db.NewExceptionIterator(0, 10)

	// make sure Release is not blocking.
	done := make(chan bool)
	go func() {
		iter.Release()
		close(done)
	}()

	select {
	case <-done:
		return
	case <-time.After(time.Second):
		t.Fatal("Release blocked unexpectedly")
	}

}

func TestExceptionIterator_DecodeSuccess(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockExceptionDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := testutil.KeyValue_Generate(nil, 70, 1, 1, 5, 3, 3)
	mockIterator := iterator.NewArrayIterator(kv)
	expected := &substate.Exception{}

	mockDb.EXPECT().NewIterator([]byte(ExceptionDBPrefix), blockTx).Return(mockIterator)
	mockDb.EXPECT().decodeToException(gomock.Any(), gomock.Any()).Return(expected, nil)

	// when
	exceptionIterator := newExceptionIterator(mockDb, blockTx)
	actual, err := exceptionIterator.decode(rawEntry{
		key:   []byte(ExceptionDBPrefix + "34567891"),
		value: []byte{1, 2, 3, 4, 5, 6, 7, 8}},
	)

	// then
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func TestExceptionIterator_DecodeFail(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockExceptionDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := testutil.KeyValue_Generate(nil, 70, 1, 1, 5, 3, 3)
	mockIterator := iterator.NewArrayIterator(kv)

	mockDb.EXPECT().NewIterator([]byte(ExceptionDBPrefix), blockTx).Return(mockIterator)

	// when
	exceptionIterator := newExceptionIterator(mockDb, blockTx)
	actual, err := exceptionIterator.decode(rawEntry{
		key:   []byte(ExceptionDBPrefix),
		value: []byte{1, 2, 3, 4, 5, 6, 7, 8}},
	)

	// then
	assert.NotNil(t, err)
	assert.Nil(t, actual)
}

func TestExceptionIterator_StartFailReading(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockExceptionDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := &testutil.KeyValue{}
	kv.PutU([]byte(ExceptionDBPrefix+"12345678"), []byte("a"))
	kv.PutU([]byte(ExceptionDBPrefix+"1234567890123457error"), []byte("b"))
	mockIterator := iterator.NewArrayIterator(kv)
	mockException := testException

	mockDb.EXPECT().NewIterator([]byte(ExceptionDBPrefix), blockTx).Return(mockIterator)
	mockDb.EXPECT().decodeToException(gomock.Any(), gomock.Any()).Return(mockException, nil)

	// when
	count := 0
	iter := newExceptionIterator(mockDb, blockTx)
	iter.start(1)
	for iter.Next() {
		tx := iter.Value()
		assert.Equal(t, mockException, tx)
		count += 1
	}
	iter.Release()
	err := iter.Error()

	// then
	assert.NotNil(t, err)
	assert.Equal(t, 1, count)
}

func TestExceptionIterator_StartFailDecoding(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockExceptionDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := &testutil.KeyValue{}
	kv.PutU([]byte(ExceptionDBPrefix+"12345678"), []byte("a"))
	kv.PutU([]byte(ExceptionDBPrefix+"90123457"), []byte("b"))
	mockIterator := iterator.NewArrayIterator(kv)
	mockException := testException
	mockError := errors.New("error")
	mockDb.EXPECT().NewIterator([]byte(ExceptionDBPrefix), blockTx).Return(mockIterator)
	gomock.InOrder(
		mockDb.EXPECT().decodeToException(gomock.Any(), gomock.Any()).Return(mockException, nil),
		mockDb.EXPECT().decodeToException(gomock.Any(), gomock.Any()).Return(nil, mockError),
	)

	var expectedValues = []*substate.Exception{mockException, nil}

	// when
	count := 0
	iter := newExceptionIterator(mockDb, blockTx)
	iter.start(1)
	for iter.Next() {
		tx := iter.Value()
		assert.Equal(t, expectedValues[count], tx)
		count += 1
	}
	iter.Release()
	err := iter.Error()

	// then
	assert.NotNil(t, err)
	assert.Equal(t, mockError.Error(), err.Error())
	assert.Equal(t, 1, count)
}

func TestExceptionIterator_StartReaderStopFirstSuccess(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockExceptionDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := &testutil.KeyValue{}
	kv.PutU([]byte(ExceptionDBPrefix+"12345678"), []byte("a"))
	kv.PutU([]byte(ExceptionDBPrefix+"90123457"), []byte("b"))
	mockIterator := iterator.NewArrayIterator(MockKeyValue{kv, 10 * time.Millisecond})
	mockException := testException

	mockDb.EXPECT().NewIterator([]byte(ExceptionDBPrefix), blockTx).Return(mockIterator)
	mockDb.EXPECT().decodeToException(gomock.Any(), gomock.Any()).Do(func(a, b interface{}) {
		time.Sleep(10 * time.Millisecond)
	}).Return(mockException, nil).Times(2)

	// when
	count := 0
	iter := newExceptionIterator(mockDb, blockTx)
	iter.start(1)
	for iter.Next() {
		tx := iter.Value()
		assert.Equal(t, mockException, tx)
		count += 1
		time.Sleep(10 * time.Millisecond)
	}
	iter.Release()
	err := iter.Error()

	// then
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, count)
}

func TestExceptionIterator_StartDecoderStopFirstSuccess(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockExceptionDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := &testutil.KeyValue{}
	for i := 0; i < 100; i++ {
		var id = uint64(12345678)
		var key = fmt.Sprintf("%v%v", ExceptionDBPrefix, id+uint64(i))
		var value = fmt.Sprintf("input%v", i)
		kv.PutU([]byte(key), []byte(value))
	}
	mockIterator := iterator.NewArrayIterator(MockKeyValue{kv, 10 * time.Millisecond})
	mockException := testException

	mockDb.EXPECT().NewIterator([]byte(ExceptionDBPrefix), blockTx).Return(mockIterator)
	mockDb.EXPECT().decodeToException(gomock.Any(), gomock.Any()).Return(mockException, nil).MinTimes(50)

	// when
	count := 0
	iter := newExceptionIterator(mockDb, blockTx)
	iter.start(1)
	for iter.Next() {
		tx := iter.Value()
		assert.Equal(t, mockException, tx)
		count += 1
		if count == 50 {
			break
		}
	}
	iter.Release()
	err := iter.Error()

	// then
	assert.Equal(t, nil, err)
	assert.Equal(t, 50, count)
}

func TestExceptionIterator_StartParallelSuccess(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockExceptionDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := &testutil.KeyValue{}
	for i := 0; i < 100; i++ {
		var id = uint64(12345678)
		var key = fmt.Sprintf("%v%v", ExceptionDBPrefix, id+uint64(i))
		var value = fmt.Sprintf("input%v", i)
		kv.PutU([]byte(key), []byte(value))
	}
	mockIterator := iterator.NewArrayIterator(MockKeyValue{kv, 10 * time.Millisecond})
	mockException := testException

	mockDb.EXPECT().NewIterator([]byte(ExceptionDBPrefix), blockTx).Return(mockIterator)
	mockDb.EXPECT().decodeToException(gomock.Any(), gomock.Any()).Return(mockException, nil).Times(100)

	// when
	count := 0
	iter := newExceptionIterator(mockDb, blockTx)
	iter.start(10)
	for iter.Next() {
		tx := iter.Value()
		assert.Equal(t, mockException, tx)
		count += 1
	}
	iter.Release()
	err := iter.Error()

	// then
	assert.Equal(t, nil, err)
	assert.Equal(t, 100, count)
}

func TestExceptionIterator_StartParallelFail(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockExceptionDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := &testutil.KeyValue{}
	for i := 0; i < 100; i++ {
		var id = uint64(12345678)
		var key = fmt.Sprintf("%v%v", ExceptionDBPrefix, id+uint64(i))
		var value = fmt.Sprintf("input%v", i)
		kv.PutU([]byte(key), []byte(value))
	}
	mockIterator := iterator.NewArrayIterator(MockKeyValue{kv, 10 * time.Millisecond})
	mockException := testException
	mockError := errors.New("error")

	mockDb.EXPECT().NewIterator([]byte(ExceptionDBPrefix), blockTx).Return(mockIterator)
	gomock.InOrder(
		mockDb.EXPECT().decodeToException(gomock.Any(), gomock.Any()).Return(mockException, nil).Times(99),
		mockDb.EXPECT().decodeToException(gomock.Any(), gomock.Any()).Return(nil, mockError).Times(1),
	)

	// when
	count := 0
	iter := newExceptionIterator(mockDb, blockTx)
	iter.start(10)
	for iter.Next() {
		tx := iter.Value()
		assert.Equal(t, mockException, tx)
		count += 1
	}
	iter.Release()
	err := iter.Error()

	// then
	assert.NotNil(t, err)
	assert.Equal(t, mockError.Error(), err.Error())
	assert.True(t, count < 100)
}

func TestExceptionIterator_StartParallelReaderStopFirstFail(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockExceptionDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := &testutil.KeyValue{}
	for i := 0; i < 100; i++ {
		var id = uint64(12345678)
		var key = fmt.Sprintf("%v%v", ExceptionDBPrefix, id+uint64(i))
		var value = fmt.Sprintf("input%v", i)
		kv.PutU([]byte(key), []byte(value))
	}
	mockIterator := iterator.NewArrayIterator(MockKeyValue{kv, 10 * time.Millisecond})
	mockException := testException
	mockError := errors.New("error")

	mockDb.EXPECT().NewIterator([]byte(ExceptionDBPrefix), blockTx).Return(mockIterator)
	gomock.InOrder(
		mockDb.EXPECT().decodeToException(gomock.Any(), gomock.Any()).Do(func(a, b interface{}) {
			time.Sleep(10 * time.Millisecond)
		}).Return(mockException, nil).Times(99),
		mockDb.EXPECT().decodeToException(gomock.Any(), gomock.Any()).Return(nil, mockError).Times(1),
	)

	// when
	count := 0
	iter := newExceptionIterator(mockDb, blockTx)
	iter.start(10)
	for iter.Next() {
		tx := iter.Value()
		assert.Equal(t, mockException, tx)
		count += 1
		time.Sleep(10 * time.Millisecond)
	}
	iter.Release()
	err := iter.Error()

	// then
	assert.NotNil(t, err)
	assert.Equal(t, mockError.Error(), err.Error())
	assert.True(t, count < 100)
}

func TestExceptionIterator_StartParallelDecoderStopFirstFail(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockExceptionDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := &testutil.KeyValue{}
	for i := 0; i < 100; i++ {
		var id = uint64(12345678)
		var key = fmt.Sprintf("%v%v", ExceptionDBPrefix, id+uint64(i))
		var value = fmt.Sprintf("input%v", i)
		kv.PutU([]byte(key), []byte(value))
	}
	mockIterator := iterator.NewArrayIterator(MockKeyValue{kv, 10 * time.Millisecond})
	mockException := testException
	mockError := errors.New("error")

	mockDb.EXPECT().NewIterator([]byte(ExceptionDBPrefix), blockTx).Return(mockIterator)
	gomock.InOrder(
		mockDb.EXPECT().decodeToException(gomock.Any(), gomock.Any()).Return(mockException, nil).Times(49),
		mockDb.EXPECT().decodeToException(gomock.Any(), gomock.Any()).Return(nil, mockError).Times(1),
		mockDb.EXPECT().decodeToException(gomock.Any(), gomock.Any()).Return(mockException, nil).AnyTimes(),
	)

	// when
	count := 0
	iter := newExceptionIterator(mockDb, blockTx)
	iter.start(10)
	for iter.Next() {
		tx := iter.Value()
		assert.Equal(t, mockException, tx)
		count += 1
		if count == 50 {
			break
		}
	}
	iter.Release()
	err := iter.Error()

	// then
	assert.NotNil(t, err)
	assert.Equal(t, mockError.Error(), err.Error())
	assert.True(t, count < 50)
}
