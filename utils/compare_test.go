package utils

import (
	"math/big"
	"testing"

	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestCompare_Identical(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	input0 := getGenericSubstate()

	gomock.InOrder(
		src.EXPECT().GetBlockSubstates(uint64(0)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
		src.EXPECT().GetBlockSubstates(uint64(1)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
		src.EXPECT().GetBlockSubstates(uint64(2)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
	)

	gomock.InOrder(
		dst.EXPECT().GetBlockSubstates(uint64(0)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
		dst.EXPECT().GetBlockSubstates(uint64(1)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
		dst.EXPECT().GetBlockSubstates(uint64(2)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
	)

	app := cli.NewApp()
	ctx := cli.NewContext(app, nil, nil)
	err := Compare(ctx, src, dst, 1, 0, 2)
	assert.NoError(t, err)
}

func TestCompare_Different(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	input0 := getGenericSubstate()

	input1 := input0.Clone()
	input1.Message.Nonce = 2

	gomock.InOrder(
		src.EXPECT().GetBlockSubstates(uint64(0)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
		src.EXPECT().GetBlockSubstates(uint64(1)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
		src.EXPECT().GetBlockSubstates(uint64(2)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
	)

	gomock.InOrder(
		dst.EXPECT().GetBlockSubstates(uint64(0)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
		dst.EXPECT().GetBlockSubstates(uint64(1)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
		dst.EXPECT().GetBlockSubstates(uint64(2)).Return(map[int]*substate.Substate{
			0: input1,
		}, nil),
	)

	app := cli.NewApp()
	ctx := cli.NewContext(app, nil, nil)
	err := Compare(ctx, src, dst, 1, 0, 2)
	assert.Error(t, err)
}

func TestCompare_MissingDst(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	input0 := getGenericSubstate()

	gomock.InOrder(
		src.EXPECT().GetBlockSubstates(uint64(0)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
		src.EXPECT().GetBlockSubstates(uint64(1)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
		src.EXPECT().GetBlockSubstates(uint64(2)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
	)

	gomock.InOrder(
		dst.EXPECT().GetBlockSubstates(uint64(0)).Return(nil, nil),
		dst.EXPECT().GetBlockSubstates(uint64(1)).Return(nil, nil),
		dst.EXPECT().GetBlockSubstates(uint64(2)).Return(nil, nil),
	)

	app := cli.NewApp()
	ctx := cli.NewContext(app, nil, nil)
	err := Compare(ctx, src, dst, 1, 0, 2)
	assert.Error(t, err)
	errWant := "target db doesn't contain substates from 0-1 onwards"
	if err.Error() != errWant {
		t.Fatalf("expected error: expected %v, got %v", errWant, err)
	}
}

func TestCompareMissingSrc(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	input0 := getGenericSubstate()

	dst.EXPECT().GetBlockSubstates(uint64(0)).Return(map[int]*substate.Substate{
		0: input0,
	}, nil)

	src.EXPECT().GetBlockSubstates(uint64(0)).Return(nil, nil)

	app := cli.NewApp()
	ctx := cli.NewContext(app, nil, nil)
	err := Compare(ctx, src, dst, 1, 0, 0)
	assert.Error(t, err)
	errWant := "source db doesn't contain substate from 0-1 onwards"
	if err.Error() != errWant {
		t.Fatalf("expected error: expected %v, got %v", errWant, err)
	}
}

func getGenericSubstate() *substate.Substate {
	return &substate.Substate{
		InputSubstate:  substate.NewWorldState(),
		OutputSubstate: substate.NewWorldState(),
		Env: &substate.Env{
			Coinbase:   types.Address{1},
			Difficulty: new(big.Int).SetUint64(1),
			GasLimit:   1,
			Number:     1,
			Timestamp:  1,
			BaseFee:    new(big.Int).SetUint64(1),
		},
		Message: substate.NewMessage(
			1,
			true,
			new(big.Int).SetUint64(1),
			1, types.Address{1},
			new(types.Address),
			new(big.Int).SetUint64(1),
			[]byte{1},
			nil,
			new(int32),
			types.AccessList{},
			new(big.Int).SetUint64(1),
			new(big.Int).SetUint64(1),
			new(big.Int).SetUint64(1),
			make([]types.Hash, 0),
			[]types.SetCodeAuthorization{},
		),
		Result:      substate.NewResult(1, types.Bloom{}, []*types.Log{}, types.Address{}, 1),
		Block:       0,
		Transaction: 1,
	}
}
