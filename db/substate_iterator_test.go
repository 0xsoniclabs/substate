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

func TestSubstateIterator_Next(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(path)
	if err != nil {
		return
	}

	iter := db.NewSubstateIterator(0, 10)
	if !iter.Next() {
		t.Fatal("next must return true")
	}

	if iter.Next() {
		t.Fatal("next must return false, all substates were extracted")
	}
}

func TestSubstateIterator_Value(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(path)
	if err != nil {
		return
	}

	testSubstatorIterator_Value(db, t)
}

func testSubstatorIterator_Value(db *substateDB, t *testing.T) {
	iter := db.NewSubstateIterator(0, 10)

	if !iter.Next() {
		t.Fatal("next must return true")
	}

	tx := iter.Value()

	if tx == nil {
		t.Fatal("iterator returned nil")
	}

	if tx.Block != 37_534_834 {
		t.Fatalf("iterator returned transaction with different block number\ngot: %v\n want: %v", tx.Block, 37_534_834)
	}

	if tx.Transaction != 1 {
		t.Fatalf("iterator returned transaction with different transaction number\ngot: %v\n want: %v", tx.Transaction, 1)
	}

}

func TestSubstateIterator_Release(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(path)
	if err != nil {
		return
	}

	iter := db.NewSubstateIterator(0, 10)

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

func TestSubstateIterator_FromBlock(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(path)
	if err != nil {
		t.Fatal(err)
	}

	test2 := *getTestSubstate("default")
	test2.Block++

	err = db.PutSubstate(&test2)
	if err != nil {
		t.Fatalf("unable to put substate: %v", err)
	}

	iter := db.NewSubstateIterator(37_534_834, 10)

	if !iter.Next() {
		t.Fatal("next must return true")
	}

	ss := iter.Value()
	if ss.Block != 37_534_834 {
		t.Fatal("incorrect block number")
	}

	counter := 1
	for iter.Next() {
		counter++
	}

	if counter != 2 {
		t.Fatal("incorrect number of substates")
	}

	iter2 := db.NewSubstateIterator(37_534_835, 10)

	if !iter2.Next() {
		t.Fatal("next must return true")
	}

	ss2 := iter2.Value()

	if ss2 == nil {
		t.Fatal("iterator returned nil")
	}

	if ss2.Block != 37_534_835 {
		t.Fatalf("iterator returned transaction with different block number\ngot: %v\n want: %v", ss.Block, 37_534_835)
	}

	if ss2.Transaction != 1 {
		t.Fatalf("iterator returned transaction with different transaction number\ngot: %v\n want: %v", ss2.Transaction, 1)
	}
}

func TestSubstateIterator_DecodeSuccess(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockSubstateDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := testutil.KeyValue_Generate(nil, 70, 1, 1, 5, 3, 3)
	mockIterator := iterator.NewArrayIterator(kv)
	expected := &substate.Substate{}

	mockDb.EXPECT().NewIterator([]byte(SubstateDBPrefix), blockTx).Return(mockIterator)
	mockDb.EXPECT().decodeToSubstate(gomock.Any(), gomock.Any(), gomock.Any()).Return(expected, nil)

	// when
	substateIterator := newSubstateIterator(mockDb, blockTx)
	actual, err := substateIterator.decode(rawEntry{
		key:   []byte(SubstateDBPrefix + "3456789123456789"),
		value: []byte{1, 2, 3, 4, 5, 6, 7, 8}},
	)

	// then
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func TestSubstateIterator_DecodeFail(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockSubstateDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := testutil.KeyValue_Generate(nil, 70, 1, 1, 5, 3, 3)
	mockIterator := iterator.NewArrayIterator(kv)

	mockDb.EXPECT().NewIterator([]byte(SubstateDBPrefix), blockTx).Return(mockIterator)

	// when
	substateIterator := newSubstateIterator(mockDb, blockTx)
	actual, err := substateIterator.decode(rawEntry{
		key:   []byte(SubstateDBPrefix),
		value: []byte{1, 2, 3, 4, 5, 6, 7, 8}},
	)

	// then
	assert.NotNil(t, err)
	assert.Nil(t, actual)
}

func TestSubstateIterator_StartFailReading(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockSubstateDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := &testutil.KeyValue{}
	kv.PutU([]byte(SubstateDBPrefix+"1234567890123456"), []byte("a"))
	kv.PutU([]byte(SubstateDBPrefix+"1234567890123457error"), []byte("b"))
	mockIterator := iterator.NewArrayIterator(kv)
	mockSubstate := getTestSubstate("protobuf")

	mockDb.EXPECT().NewIterator([]byte(SubstateDBPrefix), blockTx).Return(mockIterator)
	mockDb.EXPECT().decodeToSubstate(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockSubstate, nil)

	// when
	count := 0
	iter := newSubstateIterator(mockDb, blockTx)
	iter.start(1)
	for iter.Next() {
		tx := iter.Value()
		assert.Equal(t, mockSubstate, tx)
		count += 1
	}
	iter.Release()
	err := iter.Error()

	// then
	assert.NotNil(t, err)
	assert.Equal(t, 1, count)
}

func TestSubstateIterator_StartFailDecoding(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockSubstateDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := &testutil.KeyValue{}
	kv.PutU([]byte(SubstateDBPrefix+"1234567890123456"), []byte("a"))
	kv.PutU([]byte(SubstateDBPrefix+"1234567890123457"), []byte("b"))
	mockIterator := iterator.NewArrayIterator(kv)
	mockSubstate := getTestSubstate("protobuf")
	mockError := errors.New("error")
	mockDb.EXPECT().NewIterator([]byte(SubstateDBPrefix), blockTx).Return(mockIterator)
	gomock.InOrder(
		mockDb.EXPECT().decodeToSubstate(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockSubstate, nil),
		mockDb.EXPECT().decodeToSubstate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockError),
	)

	var expectedValues = []*substate.Substate{mockSubstate, nil}

	// when
	count := 0
	iter := newSubstateIterator(mockDb, blockTx)
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

func TestSubstateIterator_StartReaderStopFirstSuccess(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockSubstateDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := &testutil.KeyValue{}
	kv.PutU([]byte(SubstateDBPrefix+"1234567890123456"), []byte("a"))
	kv.PutU([]byte(SubstateDBPrefix+"1234567890123457"), []byte("b"))
	mockIterator := iterator.NewArrayIterator(MockKeyValue{kv, 10 * time.Millisecond})
	mockSubstate := getTestSubstate("protobuf")

	mockDb.EXPECT().NewIterator([]byte(SubstateDBPrefix), blockTx).Return(mockIterator)
	mockDb.EXPECT().decodeToSubstate(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(a, b, c interface{}) {
		time.Sleep(10 * time.Millisecond)
	}).Return(mockSubstate, nil).Times(2)

	// when
	count := 0
	iter := newSubstateIterator(mockDb, blockTx)
	iter.start(1)
	for iter.Next() {
		tx := iter.Value()
		assert.Equal(t, mockSubstate, tx)
		count += 1
		time.Sleep(10 * time.Millisecond)
	}
	iter.Release()
	err := iter.Error()

	// then
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, count)
}

func TestSubstateIterator_StartDecoderStopFirstSuccess(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockSubstateDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := &testutil.KeyValue{}
	for i := 0; i < 100; i++ {
		var id = uint64(1234567890123450)
		var key = fmt.Sprintf("%v%v", SubstateDBPrefix, id+uint64(i))
		var value = fmt.Sprintf("input%v", i)
		kv.PutU([]byte(key), []byte(value))
	}
	mockIterator := iterator.NewArrayIterator(MockKeyValue{kv, 10 * time.Millisecond})
	mockSubstate := getTestSubstate("protobuf")

	mockDb.EXPECT().NewIterator([]byte(SubstateDBPrefix), blockTx).Return(mockIterator)
	mockDb.EXPECT().decodeToSubstate(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockSubstate, nil).MinTimes(50)

	// when
	count := 0
	iter := newSubstateIterator(mockDb, blockTx)
	iter.start(1)
	for iter.Next() {
		tx := iter.Value()
		assert.Equal(t, mockSubstate, tx)
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

func TestSubstateIterator_StartParallelSuccess(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockSubstateDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := &testutil.KeyValue{}
	for i := 0; i < 100; i++ {
		var id = uint64(1234567890123450)
		var key = fmt.Sprintf("%v%v", SubstateDBPrefix, id+uint64(i))
		var value = fmt.Sprintf("input%v", i)
		kv.PutU([]byte(key), []byte(value))
	}
	mockIterator := iterator.NewArrayIterator(MockKeyValue{kv, 10 * time.Millisecond})
	mockSubstate := getTestSubstate("protobuf")

	mockDb.EXPECT().NewIterator([]byte(SubstateDBPrefix), blockTx).Return(mockIterator)
	mockDb.EXPECT().decodeToSubstate(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockSubstate, nil).Times(100)

	// when
	count := 0
	iter := newSubstateIterator(mockDb, blockTx)
	iter.start(10)
	for iter.Next() {
		tx := iter.Value()
		assert.Equal(t, mockSubstate, tx)
		count += 1
	}
	iter.Release()
	err := iter.Error()

	// then
	assert.Equal(t, nil, err)
	assert.Equal(t, 100, count)
}

func TestSubstateIterator_StartParallelFail(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockSubstateDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := &testutil.KeyValue{}
	for i := 0; i < 100; i++ {
		var id = uint64(1234567890123450)
		var key = fmt.Sprintf("%v%v", SubstateDBPrefix, id+uint64(i))
		var value = fmt.Sprintf("input%v", i)
		kv.PutU([]byte(key), []byte(value))
	}
	mockIterator := iterator.NewArrayIterator(MockKeyValue{kv, 10 * time.Millisecond})
	mockSubstate := getTestSubstate("protobuf")
	mockError := errors.New("error")

	mockDb.EXPECT().NewIterator([]byte(SubstateDBPrefix), blockTx).Return(mockIterator)
	gomock.InOrder(
		mockDb.EXPECT().decodeToSubstate(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockSubstate, nil).Times(99),
		mockDb.EXPECT().decodeToSubstate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockError).Times(1),
	)

	// when
	count := 0
	iter := newSubstateIterator(mockDb, blockTx)
	iter.start(10)
	for iter.Next() {
		tx := iter.Value()
		assert.Equal(t, mockSubstate, tx)
		count += 1
	}
	iter.Release()
	err := iter.Error()

	// then
	assert.NotNil(t, err)
	assert.Equal(t, mockError.Error(), err.Error())
	assert.True(t, count < 100)
}

func TestSubstateIterator_StartParallelReaderStopFirstFail(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockSubstateDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := &testutil.KeyValue{}
	for i := 0; i < 100; i++ {
		var id = uint64(1234567890123450)
		var key = fmt.Sprintf("%v%v", SubstateDBPrefix, id+uint64(i))
		var value = fmt.Sprintf("input%v", i)
		kv.PutU([]byte(key), []byte(value))
	}
	mockIterator := iterator.NewArrayIterator(MockKeyValue{kv, 10 * time.Millisecond})
	mockSubstate := getTestSubstate("protobuf")
	mockError := errors.New("error")

	mockDb.EXPECT().NewIterator([]byte(SubstateDBPrefix), blockTx).Return(mockIterator)
	gomock.InOrder(
		mockDb.EXPECT().decodeToSubstate(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(a, b, c interface{}) {
			time.Sleep(10 * time.Millisecond)
		}).Return(mockSubstate, nil).Times(99),
		mockDb.EXPECT().decodeToSubstate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockError).Times(1),
	)

	// when
	count := 0
	iter := newSubstateIterator(mockDb, blockTx)
	iter.start(10)
	for iter.Next() {
		tx := iter.Value()
		assert.Equal(t, mockSubstate, tx)
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

func TestSubstateIterator_StartParallelDecoderStopFirstFail(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockSubstateDB(ctrl)
	start := 10
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	kv := &testutil.KeyValue{}
	for i := 0; i < 100; i++ {
		var id = uint64(1234567890123450)
		var key = fmt.Sprintf("%v%v", SubstateDBPrefix, id+uint64(i))
		var value = fmt.Sprintf("input%v", i)
		kv.PutU([]byte(key), []byte(value))
	}
	mockIterator := iterator.NewArrayIterator(MockKeyValue{kv, 10 * time.Millisecond})
	mockSubstate := getTestSubstate("protobuf")
	mockError := errors.New("error")

	mockDb.EXPECT().NewIterator([]byte(SubstateDBPrefix), blockTx).Return(mockIterator)
	gomock.InOrder(
		mockDb.EXPECT().decodeToSubstate(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockSubstate, nil).Times(49),
		mockDb.EXPECT().decodeToSubstate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockError).Times(1),
		mockDb.EXPECT().decodeToSubstate(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockSubstate, nil).AnyTimes(),
	)

	// when
	count := 0
	iter := newSubstateIterator(mockDb, blockTx)
	iter.start(10)
	for iter.Next() {
		tx := iter.Value()
		assert.Equal(t, mockSubstate, tx)
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

type MockKeyValue struct {
	kv    *testutil.KeyValue
	delay time.Duration
}

func (kv MockKeyValue) Len() int {
	return kv.kv.Len()
}

func (kv MockKeyValue) Search(key []byte) int {
	return kv.kv.Search(key)
}

func (kv MockKeyValue) Index(i int) (key, value []byte) {
	if kv.delay > 0 {
		time.Sleep(kv.delay)
	}
	return kv.kv.Index(i)
}
