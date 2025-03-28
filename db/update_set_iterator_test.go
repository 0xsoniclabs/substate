package db

import (
	"errors"
	"testing"
	"time"

	"github.com/holiman/uint256"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/types/rlp"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/testutil"
	"go.uber.org/mock/gomock"
)

func TestUpdateSetIterator_Next(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutUpdateSet(path)
	if err != nil {
		return
	}

	iter := db.NewUpdateSetIterator(0, 10)
	if !iter.Next() {
		t.Fatal("next must return true")
	}

	if iter.Next() {
		t.Fatal("next must return false, all update-sets were extracted")
	}
}

func TestUpdateSetIterator_Value(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutUpdateSet(path)
	if err != nil {
		return
	}

	iter := db.NewUpdateSetIterator(0, 10)

	if !iter.Next() {
		t.Fatal("next must return true")
	}

	tx := iter.Value()

	if tx == nil {
		t.Fatal("iterator returned nil")
	}

	if tx.Block != 1 {
		t.Fatalf("iterator returned UpdateSet with different block number\ngot: %v\n want: %v", tx.Block, 1)
	}

}

func TestUpdateSetIterator_Release(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutUpdateSet(path)
	if err != nil {
		return
	}

	iter := db.NewUpdateSetIterator(0, 10)

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

func TestUpdateSetIterator_DecodeSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mock DB
	mockDB := NewMockUpdateDB(ctrl)

	// Create iterator
	iter := &updateSetIterator{
		db:       mockDB,
		endBlock: 100,
	}

	// Create sample block data
	blockKey := UpdateDBKey(42)

	// Create sample RLP data
	mockWorldState := updateset.UpdateSet{
		WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
		Block:           0,
		DeletedAccounts: []types.Address{},
	}

	updateSetRLP := updateset.UpdateSetRLP{
		WorldState:      mockWorldState.ToWorldStateRLP(),
		DeletedAccounts: []types.Address{},
	}

	// Encode the data
	rlpData, err := rlp.EncodeToBytes(updateSetRLP)
	assert.Nil(t, err)

	// Setup mock for GetCode
	mockDB.EXPECT().GetCode(gomock.Any()).Return([]byte("code"), nil).AnyTimes()

	// Call decode with our test data
	result, err := iter.decode(rawEntry{
		key:   blockKey,
		value: rlpData,
	})

	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestUpdateSetIterator_DecodeFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mock DB
	mockDB := NewMockUpdateDB(ctrl)

	// Create iterator
	iter := &updateSetIterator{
		db:       mockDB,
		endBlock: 100,
	}

	// Test case 1: Invalid key
	result, err := iter.decode(rawEntry{
		key:   []byte("invalid-key"),
		value: []byte("doesn't matter"),
	})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid update-set key")

	// Test case 2: Valid key but invalid RLP data
	blockNumber := uint64(42)
	blockKey := UpdateDBKey(blockNumber)
	result, err = iter.decode(rawEntry{
		key:   blockKey,
		value: []byte("invalid rlp data"),
	})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "rlp")

	// Test case 3: Valid key and RLP but ToWorldState fails
	// Create sample update set
	// Create sample RLP data
	mockWorldState := updateset.UpdateSet{
		WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
		Block:           0,
		DeletedAccounts: []types.Address{},
	}

	updateSetRLP := updateset.UpdateSetRLP{
		WorldState:      mockWorldState.ToWorldStateRLP(),
		DeletedAccounts: []types.Address{},
	}

	// Encode the data
	rlpData, err := rlp.EncodeToBytes(updateSetRLP)
	assert.Nil(t, err)

	// Setup mock for GetCode to return an error
	expectedErr := errors.New("code retrieval failed")
	mockDB.EXPECT().GetCode(gomock.Any()).Return(nil, expectedErr).AnyTimes()

	result, err = iter.decode(rawEntry{
		key:   blockKey,
		value: rlpData,
	})
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUpdateSetIterator_StartSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockUpdateDB(ctrl)

	rlpData, _ := rlp.EncodeToBytes(updateset.UpdateSetRLP{
		WorldState: updateset.UpdateSet{
			WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
			Block:           0,
			DeletedAccounts: []types.Address{},
		}.ToWorldStateRLP(),
		DeletedAccounts: []types.Address{},
	})

	kv := &testutil.KeyValue{}
	kv.PutU(UpdateDBKey(0), rlpData)
	kv.PutU(UpdateDBKey(999), rlpData)
	mockIter := iterator.NewArrayIterator(kv)
	iter := &updateSetIterator{
		genericIterator: newIterator[*updateset.UpdateSet](mockIter),
		db:              mockDB,
		endBlock:        100,
	}

	mockDB.EXPECT().GetCode(gomock.Any()).Return([]byte("code"), nil).AnyTimes()

	// Call start with a dummy value
	iter.start(0)
	iter.Release()
	err := iter.Error()
	assert.Nil(t, err)
}

func TestUpdateSetIterator_StartDecodeKeyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockUpdateDB(ctrl)

	rlpData, _ := rlp.EncodeToBytes(updateset.UpdateSetRLP{
		WorldState: updateset.UpdateSet{
			WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
			Block:           0,
			DeletedAccounts: []types.Address{},
		}.ToWorldStateRLP(),
		DeletedAccounts: []types.Address{},
	})

	kv := &testutil.KeyValue{}
	kv.PutU([]byte{1, 2, 3}, rlpData)
	mockIter := iterator.NewArrayIterator(kv)
	iter := &updateSetIterator{
		genericIterator: newIterator[*updateset.UpdateSet](mockIter),
		db:              mockDB,
		endBlock:        100,
	}

	// Call start with a dummy value
	iter.start(0)
	iter.Release()
	err := iter.Error()
	assert.NotNil(t, err)
}

func TestUpdateSetIterator_StartDecodeValueFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockUpdateDB(ctrl)

	rlpData, _ := rlp.EncodeToBytes(updateset.UpdateSetRLP{
		WorldState: updateset.UpdateSet{
			WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
			Block:           0,
			DeletedAccounts: []types.Address{},
		}.ToWorldStateRLP(),
		DeletedAccounts: []types.Address{},
	})

	kv := &testutil.KeyValue{}
	kv.PutU(UpdateDBKey(0), rlpData)
	mockIter := iterator.NewArrayIterator(kv)
	iter := &updateSetIterator{
		genericIterator: newIterator[*updateset.UpdateSet](mockIter),
		db:              mockDB,
		endBlock:        100,
	}
	expectedErr := errors.New("code retrieval failed")
	mockDB.EXPECT().GetCode(gomock.Any()).Return(nil, expectedErr).AnyTimes()

	// Call start with a dummy value
	iter.start(0)
	iter.Release()
	err := iter.Error()
	assert.NotNil(t, err)
}
