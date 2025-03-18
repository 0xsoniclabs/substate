package db

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
)

func getTestSubstate(encoding string) *substate.Substate {
	txType := int32(substate.AccessListTxType)
	ss := &substate.Substate{
		InputSubstate:  substate.NewWorldState().Add(types.Address{1}, 1, new(big.Int).SetUint64(1), nil),
		OutputSubstate: substate.NewWorldState().Add(types.Address{2}, 2, new(big.Int).SetUint64(2), nil),
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
			types.AccessList{{types.Address{1}, []types.Hash{{1}, {2}}}}, new(big.Int).SetUint64(1),
			new(big.Int).SetUint64(1), new(big.Int).SetUint64(1), make([]types.Hash, 0)),
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
	if encoding != "protobuf" {
		ss.Env.Random = nil
		ss.Message.AccessList = types.AccessList{}
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
