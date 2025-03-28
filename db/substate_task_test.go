package db

import (
	"errors"
	"testing"

	"github.com/holiman/uint256"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/mock/gomock"
)

func TestSubstateTaskPool_Execute(t *testing.T) {
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

	stPool := SubstateTaskPool{
		Name: "test",

		TaskFunc: func(block uint64, tx int, substate *substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},

		First: ts.Block,
		Last:  ts.Block + 1,

		Workers: 1,
		DB:      db,
	}

	err = stPool.Execute()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubstateTaskPool_ExecuteBlock(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	ts := getTestSubstate("default")
	stPool := SubstateTaskPool{
		Name: "test",

		TaskFunc: func(block uint64, tx int, substate *substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},

		First: ts.Block,
		Last:  ts.Block + 1,

		Workers: 1,
		DB:      db,
	}

	numTx, gas, err := stPool.ExecuteBlock(ts.Block)
	require.Nil(t, err)
	require.Equal(t, int64(1), numTx)
	require.Equal(t, ts.Message.Gas, uint64(gas))
}

func TestSubstateTaskPool_ExecuteBlock_TaskFuncErr(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	ts := getTestSubstate("default")
	stPool := SubstateTaskPool{
		Name: "test",

		TaskFunc: func(block uint64, tx int, substate *substate.Substate, taskPool *SubstateTaskPool) error {
			return errors.New("test error")
		},

		First: ts.Block,
		Last:  ts.Block + 1,

		Workers: 1,
		DB:      db,
	}

	_, _, err = stPool.ExecuteBlock(ts.Block)
	require.Error(t, err)
}

func TestSubstateTaskPool_ExecuteBlockNilTaskFunc(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	ts := getTestSubstate("default")
	stPool := SubstateTaskPool{
		Name: "test",

		BlockFunc: func(block uint64, transactions map[int]*substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},

		First: ts.Block,
		Last:  ts.Block + 1,

		Workers: 1,
		DB:      db,
	}

	numTx, gas, err := stPool.ExecuteBlock(ts.Block)
	require.Nil(t, err)
	require.Equal(t, int64(1), numTx)
	require.Equal(t, int64(0), gas)
}

func TestSubstateTaskPool_ExecuteBlockDBError(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := newSubstateDB(dbPath, nil, nil, nil)
	if err != nil {
		t.Fatalf("cannot open db; %v", err)
	}

	ts := getTestSubstate("default")
	stPool := SubstateTaskPool{
		Name: "test",

		BlockFunc: func(block uint64, transactions map[int]*substate.Substate, taskPool *SubstateTaskPool) error {
			return errors.New("test error")
		},

		First: ts.Block,
		Last:  ts.Block + 1,

		Workers: 1,
		DB:      db,
	}

	_, _, err = stPool.ExecuteBlock(ts.Block)
	require.Error(t, err)
}

func TestSubstateTaskPool_ExecuteBlockSkipTransferTx(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	ts := getTestSubstate("default")
	stPool := SubstateTaskPool{
		Name: "test",

		TaskFunc: func(block uint64, tx int, substate *substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},

		First: ts.Block,
		Last:  ts.Block + 1,

		SkipTransferTxs: true,

		Workers: 1,
		DB:      db,
	}

	numTx, gas, err := stPool.ExecuteBlock(ts.Block)
	require.Nil(t, err)
	require.Equal(t, int64(0), numTx)
	require.Equal(t, int64(0), gas)
}

func TestSubstateTaskPool_ExecuteBlockSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trans := make(map[int]*substate.Substate)
	trans[0] = getTestSubstate("default")
	trans[1] = getTestSubstate("default")

	mockDb := NewMockSubstateDB(ctrl)
	mockDb.EXPECT().GetBlockSubstates(gomock.Any()).Return(trans, nil).AnyTimes()

	task := &SubstateTaskPool{
		Name: "TestTask",
		BlockFunc: func(block uint64, transactions map[int]*substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},
		TaskFunc: func(block uint64, tx int, substate *substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},
		First:           0,
		Last:            0,
		Workers:         0,
		SkipTransferTxs: false,
		SkipCallTxs:     false,
		SkipCreateTxs:   false,
		Ctx:             nil,
		DB:              mockDb,
	}
	tx, gas, err := task.ExecuteBlock(0)

	assert.Equal(t, int64(2), tx)
	assert.Equal(t, int64(2), gas)
	assert.Nil(t, err)
}

func TestSubstateTaskPool_ExecuteBlockNoTaskFuncSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trans := make(map[int]*substate.Substate)
	trans[0] = getTestSubstate("default")
	trans[1] = getTestSubstate("default")

	mockDb := NewMockSubstateDB(ctrl)
	mockDb.EXPECT().GetBlockSubstates(gomock.Any()).Return(trans, nil).AnyTimes()

	task := &SubstateTaskPool{
		Name: "TestTask",
		BlockFunc: func(block uint64, transactions map[int]*substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},
		First:           0,
		Last:            0,
		Workers:         0,
		SkipTransferTxs: false,
		SkipCallTxs:     false,
		SkipCreateTxs:   false,
		Ctx:             nil,
		DB:              mockDb,
	}
	tx, gas, err := task.ExecuteBlock(0)

	assert.Equal(t, int64(2), tx)
	assert.Equal(t, int64(0), gas)
	assert.Nil(t, err)
}

func TestSubstateTaskPool_ExecuteBlockSkipSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trans := make(map[int]*substate.Substate)
	input := getTestSubstate("default")

	value1 := getTestSubstate("default")
	value2 := getTestSubstate("default")
	value2.InputSubstate[*input.Message.To] = substate.NewAccount(
		0, uint256.NewInt(0), []byte("code"),
	)
	value3 := getTestSubstate("default")
	value3.Message.To = nil

	trans[0] = value1
	trans[1] = value2
	trans[2] = value3

	mockDb := NewMockSubstateDB(ctrl)
	mockDb.EXPECT().GetBlockSubstates(gomock.Any()).Return(trans, nil).AnyTimes()

	// case 1: skip transfer txs
	task := &SubstateTaskPool{
		Name: "TestTask",
		BlockFunc: func(block uint64, transactions map[int]*substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},
		TaskFunc: func(block uint64, tx int, substate *substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},
		First:           0,
		Last:            0,
		Workers:         0,
		SkipTransferTxs: true,
		SkipCallTxs:     false,
		SkipCreateTxs:   false,
		Ctx:             nil,
		DB:              mockDb,
	}
	tx, gas, err := task.ExecuteBlock(0)

	assert.Equal(t, int64(2), tx)
	assert.Equal(t, int64(2), gas)
	assert.Nil(t, err)

	// case 2: skip call txs
	task = &SubstateTaskPool{
		Name: "TestTask",
		BlockFunc: func(block uint64, transactions map[int]*substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},
		TaskFunc: func(block uint64, tx int, substate *substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},
		First:           0,
		Last:            0,
		Workers:         0,
		SkipTransferTxs: false,
		SkipCallTxs:     true,
		SkipCreateTxs:   false,
		Ctx:             nil,
		DB:              mockDb,
	}
	tx, gas, err = task.ExecuteBlock(0)

	assert.Equal(t, int64(2), tx)
	assert.Equal(t, int64(2), gas)
	assert.Nil(t, err)

	// case 3: skip create txs
	task = &SubstateTaskPool{
		Name: "TestTask",
		BlockFunc: func(block uint64, transactions map[int]*substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},
		TaskFunc: func(block uint64, tx int, substate *substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},
		First:           0,
		Last:            0,
		Workers:         0,
		SkipTransferTxs: false,
		SkipCallTxs:     false,
		SkipCreateTxs:   true,
		Ctx:             nil,
		DB:              mockDb,
	}
	tx, gas, err = task.ExecuteBlock(0)

	assert.Equal(t, int64(2), tx)
	assert.Equal(t, int64(2), gas)
	assert.Nil(t, err)
}

func TestSubstateTaskPool_ExecuteBlockGetBlockSubstateFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockSubstateDB(ctrl)
	mockDb.EXPECT().GetBlockSubstates(gomock.Any()).Return(nil, leveldb.ErrClosed).AnyTimes()

	task := &SubstateTaskPool{
		Name: "TestTask",
		BlockFunc: func(block uint64, transactions map[int]*substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},
		TaskFunc: func(block uint64, tx int, substate *substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},
		First:           0,
		Last:            0,
		Workers:         0,
		SkipTransferTxs: false,
		SkipCallTxs:     false,
		SkipCreateTxs:   false,
		Ctx:             nil,
		DB:              mockDb,
	}
	tx, gas, err := task.ExecuteBlock(0)

	assert.Equal(t, int64(0), tx)
	assert.Equal(t, int64(0), gas)
	assert.NotNil(t, err)
}

func TestSubstateTaskPool_ExecuteBlockBlockFuncFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockSubstateDB(ctrl)
	trans := make(map[int]*substate.Substate)
	trans[0] = getTestSubstate("default")
	trans[1] = getTestSubstate("default")
	mockDb.EXPECT().GetBlockSubstates(gomock.Any()).Return(trans, nil).AnyTimes()

	task := &SubstateTaskPool{
		Name: "TestTask",
		BlockFunc: func(block uint64, transactions map[int]*substate.Substate, taskPool *SubstateTaskPool) error {
			return errors.New("error")
		},
		First:           0,
		Last:            0,
		Workers:         0,
		SkipTransferTxs: false,
		SkipCallTxs:     false,
		SkipCreateTxs:   false,
		Ctx:             nil,
		DB:              mockDb,
	}
	tx, gas, err := task.ExecuteBlock(0)

	assert.Equal(t, int64(0), tx)
	assert.Equal(t, int64(0), gas)
	assert.NotNil(t, err)
}

func TestSubstateTaskPool_ExecuteBlockTaskFuncFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockSubstateDB(ctrl)
	trans := make(map[int]*substate.Substate)
	trans[0] = getTestSubstate("default")
	trans[1] = getTestSubstate("default")
	mockDb.EXPECT().GetBlockSubstates(gomock.Any()).Return(trans, nil).AnyTimes()

	task := &SubstateTaskPool{
		Name: "TestTask",
		BlockFunc: func(block uint64, transactions map[int]*substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},
		TaskFunc: func(block uint64, tx int, substate *substate.Substate, taskPool *SubstateTaskPool) error {
			return errors.New("error")
		},
		First:           0,
		Last:            0,
		Workers:         0,
		SkipTransferTxs: false,
		SkipCallTxs:     false,
		SkipCreateTxs:   false,
		Ctx:             nil,
		DB:              mockDb,
	}
	tx, gas, err := task.ExecuteBlock(0)

	assert.Equal(t, int64(0), tx)
	assert.Equal(t, int64(0), gas)
	assert.NotNil(t, err)
}

func TestSubstateTaskPool_ExecuteSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockSubstateDB(ctrl)

	// Setup data for block 10
	trans1 := make(map[int]*substate.Substate)
	trans1[0] = getTestSubstate("default")
	mockDb.EXPECT().GetBlockSubstates(uint64(10)).Return(trans1, nil)
	mockDb.EXPECT().GetBlockSubstates(uint64(11)).Return(trans1, nil)

	taskPool := &SubstateTaskPool{
		Name: "TestExecuteTaskFail",
		TaskFunc: func(block uint64, tx int, s *substate.Substate, pool *SubstateTaskPool) error {
			return nil
		},
		First:   10,
		Last:    11,
		Workers: 2,
		DB:      mockDb,
	}

	// Run Execute and expect error
	err := taskPool.Execute()
	assert.Nil(t, err)
}

func TestSubstateTaskPool_ExecuteFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockSubstateDB(ctrl)

	// Setup data for block 10
	trans1 := make(map[int]*substate.Substate)
	trans1[0] = getTestSubstate("default")
	mockDb.EXPECT().GetBlockSubstates(uint64(10)).Return(trans1, nil)

	// Setup data for block 11 - will trigger error
	expectedErr := errors.New("db error")
	mockDb.EXPECT().GetBlockSubstates(uint64(11)).Return(nil, expectedErr)

	taskPool := &SubstateTaskPool{
		Name: "TestExecuteTaskFail",
		TaskFunc: func(block uint64, tx int, s *substate.Substate, pool *SubstateTaskPool) error {
			return nil
		},
		First:   10,
		Last:    11,
		Workers: 2,
		DB:      mockDb,
	}

	// Run Execute and expect error
	err := taskPool.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr.Error())
}
