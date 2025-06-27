package main

import (
	"errors"
	"flag"
	"fmt"
	"math/big"
	"testing"

	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/utils"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestRLPtoProtobufCommand_ParsingFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	set := flag.NewFlagSet("test", 0)
	_ = set.String(utils.BlockSegmentFlag.Name, "0-abc", "")
	_ = set.String(utils.WorkersFlag.Name, "1", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)

	command := rlpToProtobufCommand{
		src: src,
		dst: dst,
		ctx: ctx,
	}

	err := command.execute()
	expected := errors.New("invalid block segment string: \"0-abc\"")
	assert.Equal(t, expected, err)
}

func TestRLPtoProtobufCommand_SetEncodingFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	set := flag.NewFlagSet("test", 0)
	_ = set.String(utils.BlockSegmentFlag.Name, "0-2", "")
	_ = set.String(utils.WorkersFlag.Name, "1", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)
	mockErr := errors.New("error")

	dst.EXPECT().SetSubstateEncoding(db.ProtobufEncodingSchema).Return(mockErr)

	command := rlpToProtobufCommand{
		src: src,
		dst: dst,
		ctx: ctx,
	}

	err := command.execute()

	assert.Equal(t, mockErr, err)
}

func TestRLPtoProtobufCommand_ExecuteUpgradeFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	set := flag.NewFlagSet("test", 0)
	_ = set.String(utils.BlockSegmentFlag.Name, "0-2", "")
	_ = set.String(utils.WorkersFlag.Name, "1", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)

	mockErr := errors.New("error")
	command := rlpToProtobufCommand{
		src: src,
		dst: dst,
		ctx: ctx,
	}

	input0 := &substate.Substate{
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
			1,
			types.Address{1},
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
		Block:       37_534_834,
		Transaction: 1,
	}

	dst.EXPECT().SetSubstateEncoding(db.ProtobufEncodingSchema).Return(nil)
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

	dst.EXPECT().PutSubstate(input0).Return(nil)
	dst.EXPECT().PutSubstate(input0).Return(nil)
	dst.EXPECT().PutSubstate(input0).Return(mockErr)
	err := command.execute()

	expected := fmt.Errorf("rlp-to-protobuf: 2_0: %w", fmt.Errorf("failed to put substate: %w", mockErr))
	assert.Equal(t, expected, err)
}

func TestRLPtoProtobufCommand_ExecuteSuccessful(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	set := flag.NewFlagSet("test", 0)
	_ = set.String(utils.BlockSegmentFlag.Name, "0-2", "")
	_ = set.String(utils.WorkersFlag.Name, "1", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)

	command := rlpToProtobufCommand{
		src: src,
		dst: dst,
		ctx: ctx,
	}

	input0 := &substate.Substate{
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
		Block:       37_534_834,
		Transaction: 1,
	}

	dst.EXPECT().SetSubstateEncoding(db.ProtobufEncodingSchema).Return(nil)
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

	dst.EXPECT().PutSubstate(input0).Return(nil).Times(3)

	err := command.execute()
	assert.Nil(t, err)
}

func TestRLPtoProtobufCommand_ExecuteParallelSuccessful(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	set := flag.NewFlagSet("test", 0)
	_ = set.String(utils.BlockSegmentFlag.Name, "0-2", "")
	_ = set.String(utils.WorkersFlag.Name, "4", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)

	command := rlpToProtobufCommand{
		src: src,
		dst: dst,
		ctx: ctx,
	}

	input0 := &substate.Substate{
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
			1,
			types.Address{1},
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
		Block:       37_534_834,
		Transaction: 1,
	}

	dst.EXPECT().SetSubstateEncoding(db.ProtobufEncodingSchema).Return(nil)

	src.EXPECT().GetBlockSubstates(uint64(0)).Return(map[int]*substate.Substate{
		0: input0,
	}, nil)
	src.EXPECT().GetBlockSubstates(uint64(1)).Return(map[int]*substate.Substate{
		0: input0,
	}, nil)
	src.EXPECT().GetBlockSubstates(uint64(2)).Return(map[int]*substate.Substate{
		0: input0,
	}, nil)

	dst.EXPECT().PutSubstate(input0).Return(nil).Times(3)

	err := command.execute()
	assert.Nil(t, err)
}

func TestRLPtoProtobufCommand_ExecuteFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	set := flag.NewFlagSet("test", 0)
	_ = set.String(utils.BlockSegmentFlag.Name, "0-2", "")
	_ = set.String(utils.WorkersFlag.Name, "1", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)
	mockErr := errors.New("error")

	command := rlpToProtobufCommand{
		src: src,
		dst: dst,
		ctx: ctx,
	}

	input0 := &substate.Substate{
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
			1,
			types.Address{1},
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
		Block:       37_534_834,
		Transaction: 1,
	}

	dst.EXPECT().SetSubstateEncoding(db.ProtobufEncodingSchema).Return(nil)
	gomock.InOrder(
		src.EXPECT().GetBlockSubstates(uint64(0)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
		src.EXPECT().GetBlockSubstates(uint64(1)).Return(nil, mockErr),
		src.EXPECT().GetBlockSubstates(uint64(2)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
	)

	dst.EXPECT().PutSubstate(input0).Return(nil).Times(2)

	err := command.execute()
	assert.Equal(t, mockErr, err)
}

func TestRLPtoProtobufCommand_ExecuteParallelFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	set := flag.NewFlagSet("test", 0)
	_ = set.String(utils.BlockSegmentFlag.Name, "0-2", "")
	_ = set.String(utils.WorkersFlag.Name, "4", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)
	mockErr := errors.New("error")

	command := rlpToProtobufCommand{
		src: src,
		dst: dst,
		ctx: ctx,
	}

	input0 := &substate.Substate{
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
			1,
			types.Address{1},
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
		Block:       37_534_834,
		Transaction: 1,
	}

	dst.EXPECT().SetSubstateEncoding(db.ProtobufEncodingSchema).Return(nil)

	src.EXPECT().GetBlockSubstates(uint64(0)).Return(map[int]*substate.Substate{
		0: input0,
	}, nil)
	src.EXPECT().GetBlockSubstates(uint64(1)).Return(nil, mockErr)
	src.EXPECT().GetBlockSubstates(uint64(2)).Return(map[int]*substate.Substate{
		0: input0,
	}, nil)

	dst.EXPECT().PutSubstate(input0).Return(nil).Times(2)

	err := command.execute()
	assert.Equal(t, mockErr, err)
}

func TestTestingOfFunctionFromSeparatePackage(t *testing.T) {
	list := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	target := 8
	db.GenerateBinarySearchInList(list, target)
}
