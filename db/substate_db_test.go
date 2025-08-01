package db

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"

	"github.com/holiman/uint256"

	"github.com/0xsoniclabs/substate/protobuf"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/testutil"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
)

func getTestSubstate(encoding SubstateEncodingSchema) *substate.Substate {
	txType := int32(substate.SetCodeTxType)
	ss := &substate.Substate{
		InputSubstate:  substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
		OutputSubstate: substate.NewWorldState().Add(types.Address{2}, 2, new(uint256.Int).SetUint64(2), nil),
		Env: &substate.Env{
			Coinbase:   types.Address{1},
			Difficulty: new(big.Int).SetUint64(1),
			GasLimit:   1,
			Number:     1,
			Timestamp:  1,
			BaseFee:    new(big.Int).SetUint64(1),
			Random:     &types.Hash{1},
		},
		Message: substate.NewMessage(
			1,
			true,
			new(big.Int).SetUint64(1),
			1,
			types.Address{1},
			new(types.Address), new(big.Int).SetUint64(1), []byte{1}, nil, &txType,
			types.AccessList{{Address: types.Address{1}, StorageKeys: []types.Hash{{1}, {2}}}}, new(big.Int).SetUint64(1),
			new(big.Int).SetUint64(1), new(big.Int).SetUint64(1), make([]types.Hash, 0),
			[]types.SetCodeAuthorization{
				{ChainID: *uint256.NewInt(1), Address: types.Address{1}, Nonce: 1, V: 1, R: *uint256.NewInt(1), S: *uint256.NewInt(1)},
			}),
		Result: substate.NewResult(1, types.Bloom{1}, []*types.Log{
			{
				Address: types.Address{1},
				Topics:  []types.Hash{{1}, {2}},
				Data:    []byte{1, 2, 3},
				// intentionally skipped: BlockNumber, TxIndex, Index - because protobuf Substate encoding doesn't use these values
				TxHash:    types.Hash{1},
				BlockHash: types.Hash{1},
				Removed:   false,
			},
		},
			// intentionally skipped: ContractAddress - because protobuf Substate encoding doesn't use this value,
			// instead the ContractAddress is derived from Message.From and Message.Nonce
			types.Address{},
			1),
		Block:       37_534_834,
		Transaction: 1,
	}

	// remove fields that are not supported in rlp encoding
	// TODO once protobuf becomes default add ' && encoding != "default" ' to the condition
	if encoding != ProtobufEncodingSchema {
		ss.Env.Random = nil
		ss.Message.AccessList = types.AccessList{}
		ss.Message.SetCodeAuthorizations = nil
	}
	return ss
}

func TestSubstateDB_PutSubstate(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
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

func TestSubstateDB_HasSubstate(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	has, err := db.HasSubstate(37_534_834, 1)
	if err != nil {
		t.Fatalf("has substate returned error; %v", err)
	}

	if !has {
		t.Fatal("substate is not within db")
	}
}
func TestSubstateDB_GetSubstate(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	err = testSubstateDB_GetSubstate(db, *getTestSubstate("default"))
	if err != nil {
		t.Fatal(err)
	}
}

func testSubstateDB_GetSubstate(db *substateDB, want substate.Substate) error {
	ss, err := db.GetSubstate(37_534_834, 1)
	if err != nil {
		return fmt.Errorf("get substate returned error; %v", err)
	}

	if ss == nil {
		return errors.New("substate is nil")
	}

	if err = want.Equal(ss); err != nil {
		return fmt.Errorf("substates are different; %v", err)
	}
	return nil
}

func TestSubstateDB_DeleteSubstate(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	err = db.DeleteSubstate(37_534_834, 1)
	if err != nil {
		t.Fatalf("delete substate returned error; %v", err)
	}

	ss, err := db.GetSubstate(37_534_834, 1)
	if err == nil {
		t.Fatal("get substate must fail")
	}

	if got, want := err, leveldb.ErrNotFound; !errors.Is(got, want) {
		t.Fatalf("unexpected err, got: %v, want: %v", got, want)
	}

	if ss != nil {
		t.Fatal("substate was not deleted")
	}
}

func TestSubstateDB_getLastBlock(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	ts := getTestSubstate("default")
	// add one more substate
	if err = addSubstate(db, ts.Block+1); err != nil {
		t.Fatal(err)
	}

	block, err := db.getLastBlock()
	if err != nil {
		t.Fatal(err)
	}

	if block != 37534835 {
		t.Fatalf("incorrect block number\ngot: %v\nwant: %v", block, ts.Block+1)
	}

}

func TestSubstateDB_GetFirstSubstate(t *testing.T) {
	// save data for comparison
	want := *getTestSubstate("default")
	want.Block = 1

	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	// add one more substate
	if err = addSubstate(db, 2); err != nil {
		t.Fatal(err)
	}

	got := db.GetFirstSubstate()

	if err = (&want).Equal(got); err != nil {
		t.Fatalf("substates are different\nerr: %v\ngot: %s\nwant: %s", err, got, &want)
	}

}

func TestSubstateDB_GetLastSubstate(t *testing.T) {
	// save data for comparison
	want := *getTestSubstate("default")
	want.Block = 2

	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	// add one more substate
	if err = addSubstate(db, 2); err != nil {
		t.Fatal(err)
	}

	got, err := db.GetLastSubstate()
	if err != nil {
		t.Fatal(err)
	}

	if err = (&want).Equal(got); err != nil {
		t.Fatalf("substates are different\nerr: %v\ngot: %s\nwant: %s", err, got, &want)
	}

}
func createDbAndPutSubstate(dbPath string) (*substateDB, error) {
	db, err := newSubstateDB(dbPath, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot open db; %v", err)
	}

	if err = addSubstate(db, getTestSubstate("default").Block); err != nil {
		return nil, err
	}

	return db, nil
}

func addSubstate(db *substateDB, blk uint64) error {
	return addCustomSubstate(db, blk, getTestSubstate("default"))
}

func addCustomSubstate(db *substateDB, blk uint64, ss *substate.Substate) error {
	s := *ss
	s.Block = blk

	return db.PutSubstate(&s)
}

func TestSubstateDB_GetFirstSubstateSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)

	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}
	err := db.SetSubstateEncoding(ProtobufEncodingSchema)
	if err != nil {
		t.Fatal(err)
	}

	// case 1: return substate
	kv := &testutil.KeyValue{}
	input := getSubstate()
	encoded, err := protobuf.Encode(input, 1, 1)
	assert.Nil(t, err)
	kv.PutU(SubstateDBKey(1, 1), encoded)
	mockIter := iterator.NewArrayIterator(kv)
	mockDb.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(mockIter)
	mockDb.EXPECT().GetCode(gomock.Any()).Return([]byte("code"), nil).AnyTimes()
	output := db.GetFirstSubstate()
	assert.NotNil(t, output)

	// case 2: return nil
	kv = &testutil.KeyValue{}
	mockIter = iterator.NewArrayIterator(kv)
	mockDb.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(mockIter)
	mockDb.EXPECT().GetCode(gomock.Any()).Return([]byte("code"), nil).AnyTimes()
	output = db.GetFirstSubstate()
	assert.Nil(t, output)
}

func TestSubstateDB_GetFirstSubstateFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)

	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}
	err := db.SetSubstateEncoding(ProtobufEncodingSchema)
	if err != nil {
		t.Fatal(err)
	}

	kv := &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte{2})
	mockIter := iterator.NewArrayIterator(kv)
	mockDb.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(mockIter)
	mockDb.EXPECT().GetCode(gomock.Any()).Return([]byte("code"), nil).AnyTimes()
	output := db.GetFirstSubstate()
	assert.Nil(t, output)
}
func TestSubstateDB_HasSubstateSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}

	mockDb.EXPECT().Has(SubstateDBKey(1, 1)).Return(true, nil)
	has, err := db.HasSubstate(1, 1)

	assert.Nil(t, err)
	assert.True(t, has)
}

func TestSubstateDB_HasSubstateFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}

	mockDb.EXPECT().Has(SubstateDBKey(1, 1)).Return(false, errors.New("db error"))
	has, err := db.HasSubstate(1, 1)

	assert.NotNil(t, err)
	assert.False(t, has)
}

func TestSubstateDB_GetSubstateSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}
	err := db.SetSubstateEncoding(ProtobufEncodingSchema)
	if err != nil {
		t.Fatal(err)
	}

	ss := getSubstate()
	encoded, _ := db.encodeSubstate(ss, 1, 1)

	mockDb.EXPECT().GetCode(gomock.Any()).Return([]byte("code"), nil).AnyTimes()
	mockDb.EXPECT().Get(SubstateDBKey(1, 1)).Return(encoded, nil)
	result, err := db.GetSubstate(1, 1)

	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestSubstateDB_GetSubstateFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}

	// Case 1: Get error
	mockDb.EXPECT().Get(SubstateDBKey(1, 1)).Return(nil, errors.New("db error"))
	result, err := db.GetSubstate(1, 1)

	assert.NotNil(t, err)
	assert.Nil(t, result)

	// Case 2: Decode error
	err = db.SetSubstateEncoding(ProtobufEncodingSchema)
	if err != nil {
		t.Fatal(err)
	}
	mockDb.EXPECT().Get(SubstateDBKey(1, 1)).Return([]byte{1, 2, 3}, nil)
	result, err = db.GetSubstate(1, 1)

	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestSubstateDB_GetBlockSubstatesSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}
	err := db.SetSubstateEncoding(ProtobufEncodingSchema)
	if err != nil {
		t.Fatal(err)
	}

	ss := getSubstate()
	encoded, _ := db.encodeSubstate(ss, 1, 1)

	kv := &testutil.KeyValue{}
	kv.PutU(SubstateDBKey(1, 1), encoded)
	mockIter := iterator.NewArrayIterator(kv)

	mockDb.EXPECT().GetCode(gomock.Any()).Return([]byte("code"), nil).AnyTimes()
	mockDb.EXPECT().newIterator(gomock.Any()).Return(mockIter)

	substates, err := db.GetBlockSubstates(1)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(substates))
	assert.NotNil(t, substates[1])
}

func TestSubstateDB_GetBlockSubstatesFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB: mockDb,
		encoding: &substateEncoding{
			schema: ProtobufEncodingSchema,
			decode: func(bytes []byte, block uint64, tx int) (*substate.Substate, error) {
				return nil, errors.New("mock Error")
			},
			encode: protobuf.Encode,
		},
	}

	// Case 1: Invalid key
	kv := &testutil.KeyValue{}
	kv.PutU([]byte("invalid key"), []byte{1, 2, 3})
	mockIter := iterator.NewArrayIterator(kv)

	mockDb.EXPECT().newIterator(gomock.Any()).Return(mockIter)

	substates, err := db.GetBlockSubstates(1)

	assert.NotNil(t, err)
	assert.Nil(t, substates)

	// Case 2: Invalid block number
	kv = &testutil.KeyValue{}
	kv.PutU(SubstateDBKey(1, 1), []byte{1, 2, 3})
	mockIter = iterator.NewArrayIterator(kv)

	mockDb.EXPECT().GetCode(gomock.Any()).Return([]byte("code"), nil).AnyTimes()
	mockDb.EXPECT().newIterator(gomock.Any()).Return(mockIter)
	substates, err = db.GetBlockSubstates(2)
	assert.NotNil(t, err)
	assert.Nil(t, substates)

	// Case 3: Decode error
	kv = &testutil.KeyValue{}
	kv.PutU(SubstateDBKey(1, 1), []byte{1, 2, 3})
	mockIter = iterator.NewArrayIterator(kv)
	mockDb.EXPECT().GetCode(gomock.Any()).Return([]byte("code"), nil).AnyTimes()
	mockDb.EXPECT().newIterator(gomock.Any()).Return(mockIter)
	substates, err = db.GetBlockSubstates(1)
	assert.NotNil(t, err)
	assert.Nil(t, substates)

	// Case 4: Iterator error
	kv = &testutil.KeyValue{}
	kv.PutU([]byte{1, 2}, []byte{1, 2, 3})
	mockIter = iterator.NewEmptyIterator(errors.New("mock Error"))
	mockDb.EXPECT().GetCode(gomock.Any()).Return([]byte("code"), nil).AnyTimes()
	mockDb.EXPECT().newIterator(gomock.Any()).Return(mockIter)
	substates, err = db.GetBlockSubstates(1)
	assert.NotNil(t, err)
	assert.Nil(t, substates)

}

func TestSubstateDB_PutSubstateSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}
	err := db.SetSubstateEncoding(ProtobufEncodingSchema)
	if err != nil {
		t.Fatal(err)
	}

	ss := getSubstate()
	ss.Message.To = nil

	mockDb.EXPECT().PutCode(gomock.Any()).Return(nil).Times(3)
	mockDb.EXPECT().Put(SubstateDBKey(1, 1), gomock.Any()).Return(nil)

	err = db.PutSubstate(ss)
	assert.Nil(t, err)
}

func TestSubstateDB_PutSubstateFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}
	err := db.SetSubstateEncoding(ProtobufEncodingSchema)
	if err != nil {
		t.Fatal(err)
	}

	ss := getSubstate()
	ss.Message.To = nil

	// Case 1: Input code save error
	mockDb.EXPECT().PutCode(gomock.Any()).Return(errors.New("code save error"))

	err = db.PutSubstate(ss)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "cannot put preState code")

	// Case 2: Output code save error
	mockDb.EXPECT().PutCode(gomock.Any()).Return(nil).Times(1)
	mockDb.EXPECT().PutCode(gomock.Any()).Return(errors.New("code save error"))

	err = db.PutSubstate(ss)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "cannot put postState code")

	// Case 3: Message code save error
	mockDb.EXPECT().PutCode(gomock.Any()).Return(nil).Times(2)
	mockDb.EXPECT().PutCode(gomock.Any()).Return(errors.New("code save error"))

	err = db.PutSubstate(ss)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "cannot put input data")

	// Case 4: Put error
	mockDb.EXPECT().PutCode(gomock.Any()).Return(nil).Times(3)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(errors.New("put error"))

	err = db.PutSubstate(ss)

	assert.NotNil(t, err)

	// Case 5: encode error
	db = &substateDB{
		CodeDB: mockDb,
		encoding: &substateEncoding{
			schema: ProtobufEncodingSchema,
			decode: func(bytes []byte, block uint64, tx int) (*substate.Substate, error) {
				return nil, errors.New("mock Error")
			},
			encode: func(s *substate.Substate, u uint64, i int) ([]byte, error) {
				return nil, errors.New("mock Error")
			},
		},
	}
	mockDb.EXPECT().PutCode(gomock.Any()).Return(nil).Times(3)
	err = db.PutSubstate(ss)
	assert.NotNil(t, err)
}

func TestSubstateDB_DeleteSubstateSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}

	mockDb.EXPECT().Delete(SubstateDBKey(1, 1)).Return(nil)

	err := db.DeleteSubstate(1, 1)

	assert.Nil(t, err)
}

func TestSubstateDB_DeleteSubstateFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}

	mockDb.EXPECT().Delete(SubstateDBKey(1, 1)).Return(errors.New("delete error"))

	err := db.DeleteSubstate(1, 1)

	assert.NotNil(t, err)
	assert.Equal(t, "delete error", err.Error())
}

func TestSubstateDB_NewSubstateIterator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}
	err := db.SetSubstateEncoding(ProtobufEncodingSchema)
	if err != nil {
		t.Fatal(err)
	}

	// Test with empty DB
	kv := &testutil.KeyValue{}
	mockIter := iterator.NewArrayIterator(kv)

	mockDb.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(mockIter)

	iter := db.NewSubstateIterator(0, 1)

	assert.NotNil(t, iter)
	assert.False(t, iter.Next())
}

func TestSubstateDB_NewSubstateTaskPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}

	mockCtx := &cli.Context{}

	taskPool := db.NewSubstateTaskPool("test", func(block uint64, tx int, substate *substate.Substate, taskPool *SubstateTaskPool) error {
		return nil
	}, 1, 10, mockCtx)

	assert.NotNil(t, taskPool)
	assert.Equal(t, "test", taskPool.Name)
	assert.Equal(t, uint64(1), taskPool.First)
	assert.Equal(t, uint64(10), taskPool.Last)
}

func TestSubstateDB_GetLongestEncodedKeyZeroPrefixLengthSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}

	// Case: Found at index 2
	mockDb.EXPECT().hasKeyValuesFor([]byte(SubstateDBPrefix), gomock.Any()).Return(false).Times(2)
	mockDb.EXPECT().hasKeyValuesFor([]byte(SubstateDBPrefix), gomock.Any()).Return(true)

	result, err := db.getLongestEncodedKeyZeroPrefixLength()

	assert.Nil(t, err)
	assert.Equal(t, byte(2), result)
}

func TestSubstateDB_GetLongestEncodedKeyZeroPrefixLengthFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}

	// Case: Not found
	mockDb.EXPECT().hasKeyValuesFor([]byte(SubstateDBPrefix), gomock.Any()).Return(false).Times(8)

	result, err := db.getLongestEncodedKeyZeroPrefixLength()

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unable to find prefix")
	assert.Equal(t, byte(0), result)
}

func TestSubstateDB_GetLastBlockSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}

	// Case 1: zero bytes
	mockDb.EXPECT().hasKeyValuesFor(gomock.Any(), gomock.Any()).Return(true)
	mockDb.EXPECT().binarySearchForLastPrefixKey(gomock.Any()).Return(byte(0), nil).AnyTimes()

	block, err := db.getLastBlock()

	assert.Nil(t, err)
	assert.Equal(t, uint64(0x0), block)

	// Case 2: non-zero bytes
	mockDb.EXPECT().hasKeyValuesFor(gomock.Any(), gomock.Any()).Return(false).Times(2)
	mockDb.EXPECT().hasKeyValuesFor(gomock.Any(), gomock.Any()).Return(true)
	mockDb.EXPECT().binarySearchForLastPrefixKey(gomock.Any()).Return(byte(0), nil).AnyTimes()

	block, err = db.getLastBlock()

	assert.Nil(t, err)
	assert.Equal(t, uint64(0x0), block)
}

func TestSubstateDB_GetLastBlockFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}

	// Case 1: getLongestEncodedKeyZeroPrefixLength error
	mockDb.EXPECT().hasKeyValuesFor(gomock.Any(), gomock.Any()).Return(false).Times(8)
	block, err := db.getLastBlock()

	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), block)

	// Case 2: binarySearchForLastPrefixKey error
	mockDb.EXPECT().hasKeyValuesFor(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
	mockDb.EXPECT().binarySearchForLastPrefixKey(gomock.Any()).Return(byte(0), errors.New("search error"))

	block, err = db.getLastBlock()

	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), block)
}

func TestSubstateDB_GetLastSubstateSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}
	err := db.SetSubstateEncoding(ProtobufEncodingSchema)
	if err != nil {
		t.Fatal(err)
	}

	ss := getSubstate()
	ss.Block = uint64(72340172838076673)
	encoded, _ := db.encodeSubstate(ss, uint64(72340172838076673), 1)

	// Set up mock expectations for getLastBlock and GetBlockSubstates
	mockDb.EXPECT().hasKeyValuesFor(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
	mockDb.EXPECT().binarySearchForLastPrefixKey(gomock.Any()).Return(byte(1), nil).AnyTimes()

	// Mock the iteration over block substates
	kv := &testutil.KeyValue{}
	kv.PutU(SubstateDBKey(uint64(72340172838076673), 1), encoded)
	mockIter := iterator.NewArrayIterator(kv)
	mockDb.EXPECT().newIterator(gomock.Any()).Return(mockIter)
	mockDb.EXPECT().GetCode(gomock.Any()).Return([]byte("code"), nil).AnyTimes()

	// Call the method under test
	result, err := db.GetLastSubstate()

	// Verify results
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, ss.Block, result.Block)
	assert.Equal(t, ss.Transaction, result.Transaction)
}

func TestSubstateDB_GetLastSubstateFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}
	err := db.SetSubstateEncoding(ProtobufEncodingSchema)
	if err != nil {
		t.Fatal(err)
	}

	// Case 1: getLastBlock error
	mockDb.EXPECT().hasKeyValuesFor(gomock.Any(), gomock.Any()).Return(false).Times(8)
	result, err := db.GetLastSubstate()

	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unable to find prefix")

	// Case 2: GetBlockSubstates error
	mockDb.EXPECT().hasKeyValuesFor(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
	mockDb.EXPECT().binarySearchForLastPrefixKey(gomock.Any()).Return(byte(10), nil).AnyTimes()

	// Return an error from GetBlockSubstates by making the iterator return an error
	mockIter := iterator.NewEmptyIterator(errors.New("iterator error"))
	mockDb.EXPECT().newIterator(gomock.Any()).Return(mockIter)

	result, err = db.GetLastSubstate()

	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "iterator error")

	// Case 3: No substates found for last block
	mockDb.EXPECT().binarySearchForLastPrefixKey(gomock.Any()).Return(byte(10), nil).AnyTimes()

	// Return an empty result (no substates)
	emptyKv := &testutil.KeyValue{}
	mockIter = iterator.NewArrayIterator(emptyKv)
	mockDb.EXPECT().newIterator(gomock.Any()).Return(mockIter)

	result, err = db.GetLastSubstate()

	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "doesn't have any substates")
}

func TestDecodeSubstateDBKey_Success(t *testing.T) {
	key := []byte(SubstateDBPrefix + "0123456789abcdef")
	block, trans, err := DecodeSubstateDBKey(key)

	assert.Nil(t, err)
	assert.Equal(t, uint64(0x3031323334353637), block)
	assert.Equal(t, 4051376414998685030, trans)
}

func TestDecodeSubstateDBKey_Error(t *testing.T) {

	// case 1: key length is less than 18
	key := []byte(SubstateDBPrefix + "0123456789abcde")
	block, trans, err := DecodeSubstateDBKey(key)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid length")
	assert.Equal(t, uint64(0), block)
	assert.Equal(t, 0, trans)

	// case 2: prefix is not SubstateDBPrefix
	key = []byte("0123456789abcdefgh")
	block, trans, err = DecodeSubstateDBKey(key)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid prefix")
	assert.Equal(t, uint64(0), block)
	assert.Equal(t, 0, trans)
}

func TestSubstateDB_setEncoding(t *testing.T) {
	tests := []struct {
		name     string
		encoding SubstateEncodingSchema
	}{
		{
			name:     "ProtobufEncodingSchema",
			encoding: ProtobufEncodingSchema,
		},
		{
			name:     "RLPEncodingSchema",
			encoding: RLPEncodingSchema,
		},
		{
			name:     "LegacyProtobufEncodingSchema",
			encoding: LegacyProtobufEncodingAlias,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := t.TempDir()
			ss := getTestSubstate(test.encoding)
			sdb, err := NewDefaultSubstateDB(path)
			require.NoError(t, err)
			err = sdb.SetSubstateEncoding(test.encoding)
			require.NoError(t, err)
			err = sdb.PutSubstate(ss)
			require.NoError(t, err)
			require.NoError(t, sdb.Close())
			cdb, err := NewDefaultCodeDB(path)
			require.NoError(t, err)
			db := &substateDB{CodeDB: cdb}
			require.NoError(t, err)
			err = db.findAndSetEncoding()
			require.NoError(t, err)
			got := db.GetFirstSubstate()
			require.Equal(t, ss.Block, got.Block)
			require.Equal(t, ss.Transaction, got.Transaction)
		})
	}
}
