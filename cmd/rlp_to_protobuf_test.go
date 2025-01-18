package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"math/big"
	"testing"
)
import "go.uber.org/mock/gomock"

func TestRLPtoProtobufCommand_ParsingFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	set := flag.NewFlagSet("test", 0)
	_ = set.String(BlockSegmentFlag.Name, "0-abc", "")
	_ = set.String(WorkersFlag.Name, "1", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)

	command := RLPtoProtobufCommand{
		src: src,
		dst: dst,
		ctx: ctx,
	}

	err := command.Execute()
	expected := errors.New("invalid block segment string: \"0-abc\"")
	assert.Equal(t, expected, err)
}

func TestRLPtoProtobufCommand_SetEncodingFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	set := flag.NewFlagSet("test", 0)
	_ = set.String(BlockSegmentFlag.Name, "0-2", "")
	_ = set.String(WorkersFlag.Name, "1", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)

	dst.EXPECT().SetSubstateEncoding("protobuf").Return(nil, errors.New("error"))

	command := RLPtoProtobufCommand{
		src: src,
		dst: dst,
		ctx: ctx,
	}

	err := command.Execute()
	expected := errors.New("error")
	assert.Equal(t, expected, err)
}

func TestRLPtoProtobufCommand_ExecuteUpgradeFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	set := flag.NewFlagSet("test", 0)
	_ = set.String(BlockSegmentFlag.Name, "0-2", "")
	_ = set.String(WorkersFlag.Name, "1", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)

	command := RLPtoProtobufCommand{
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
		Message:     substate.NewMessage(1, true, new(big.Int).SetUint64(1), 1, types.Address{1}, new(types.Address), new(big.Int).SetUint64(1), []byte{1}, nil, types.AccessList{}, new(big.Int).SetUint64(1), new(big.Int).SetUint64(1), new(big.Int).SetUint64(1), make([]types.Hash, 0)),
		Result:      substate.NewResult(1, types.Bloom{}, []*types.Log{}, types.Address{}, 1),
		Block:       37_534_834,
		Transaction: 1,
	}

	dst.EXPECT().SetSubstateEncoding("protobuf").Return(&db.CSubstateDB{}, nil)
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
	dst.EXPECT().PutSubstate(input0).Return(errors.New("error"))
	err := command.Execute()

	expected := fmt.Errorf("rlp-to-protobuf: 2_0: %w", errors.New("error"))
	assert.Equal(t, expected, err)
}

func TestRLPtoProtobufCommand_ExecuteSuccessful(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	set := flag.NewFlagSet("test", 0)
	_ = set.String(BlockSegmentFlag.Name, "0-2", "")
	_ = set.String(WorkersFlag.Name, "1", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)

	command := RLPtoProtobufCommand{
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
		Message:     substate.NewMessage(1, true, new(big.Int).SetUint64(1), 1, types.Address{1}, new(types.Address), new(big.Int).SetUint64(1), []byte{1}, nil, types.AccessList{}, new(big.Int).SetUint64(1), new(big.Int).SetUint64(1), new(big.Int).SetUint64(1), make([]types.Hash, 0)),
		Result:      substate.NewResult(1, types.Bloom{}, []*types.Log{}, types.Address{}, 1),
		Block:       37_534_834,
		Transaction: 1,
	}

	dst.EXPECT().SetSubstateEncoding("protobuf").Return(&db.CSubstateDB{}, nil)
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

	err := command.Execute()
	assert.Nil(t, err)
}

func TestRLPtoProtobufCommand_ExecuteParallelSuccessful(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	set := flag.NewFlagSet("test", 0)
	_ = set.String(BlockSegmentFlag.Name, "0-2", "")
	_ = set.String(WorkersFlag.Name, "4", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)

	command := RLPtoProtobufCommand{
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
		Message:     substate.NewMessage(1, true, new(big.Int).SetUint64(1), 1, types.Address{1}, new(types.Address), new(big.Int).SetUint64(1), []byte{1}, nil, types.AccessList{}, new(big.Int).SetUint64(1), new(big.Int).SetUint64(1), new(big.Int).SetUint64(1), make([]types.Hash, 0)),
		Result:      substate.NewResult(1, types.Bloom{}, []*types.Log{}, types.Address{}, 1),
		Block:       37_534_834,
		Transaction: 1,
	}

	dst.EXPECT().SetSubstateEncoding("protobuf").Return(&db.CSubstateDB{}, nil)

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

	err := command.Execute()
	assert.Nil(t, err)
}

func TestRLPtoProtobufCommand_ExecuteFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	set := flag.NewFlagSet("test", 0)
	_ = set.String(BlockSegmentFlag.Name, "0-2", "")
	_ = set.String(WorkersFlag.Name, "1", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)

	command := RLPtoProtobufCommand{
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
		Message:     substate.NewMessage(1, true, new(big.Int).SetUint64(1), 1, types.Address{1}, new(types.Address), new(big.Int).SetUint64(1), []byte{1}, nil, types.AccessList{}, new(big.Int).SetUint64(1), new(big.Int).SetUint64(1), new(big.Int).SetUint64(1), make([]types.Hash, 0)),
		Result:      substate.NewResult(1, types.Bloom{}, []*types.Log{}, types.Address{}, 1),
		Block:       37_534_834,
		Transaction: 1,
	}

	dst.EXPECT().SetSubstateEncoding("protobuf").Return(&db.CSubstateDB{}, nil)
	gomock.InOrder(
		src.EXPECT().GetBlockSubstates(uint64(0)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
		src.EXPECT().GetBlockSubstates(uint64(1)).Return(nil, errors.New("error")),
		src.EXPECT().GetBlockSubstates(uint64(2)).Return(map[int]*substate.Substate{
			0: input0,
		}, nil),
	)

	dst.EXPECT().PutSubstate(input0).Return(nil).Times(2)

	err := command.Execute()
	assert.Equal(t, errors.New("error"), err)
}

func TestRLPtoProtobufCommand_ExecuteParallelFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	src := db.NewMockSubstateDB(ctrl)
	dst := db.NewMockSubstateDB(ctrl)

	set := flag.NewFlagSet("test", 0)
	_ = set.String(BlockSegmentFlag.Name, "0-2", "")
	_ = set.String(WorkersFlag.Name, "4", "")
	ctx := cli.NewContext(&cli.App{}, set, nil)

	command := RLPtoProtobufCommand{
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
		Message:     substate.NewMessage(1, true, new(big.Int).SetUint64(1), 1, types.Address{1}, new(types.Address), new(big.Int).SetUint64(1), []byte{1}, nil, types.AccessList{}, new(big.Int).SetUint64(1), new(big.Int).SetUint64(1), new(big.Int).SetUint64(1), make([]types.Hash, 0)),
		Result:      substate.NewResult(1, types.Bloom{}, []*types.Log{}, types.Address{}, 1),
		Block:       37_534_834,
		Transaction: 1,
	}

	dst.EXPECT().SetSubstateEncoding("protobuf").Return(&db.CSubstateDB{}, nil)

	src.EXPECT().GetBlockSubstates(uint64(0)).Return(map[int]*substate.Substate{
		0: input0,
	}, nil)
	src.EXPECT().GetBlockSubstates(uint64(1)).Return(nil, errors.New("error"))
	src.EXPECT().GetBlockSubstates(uint64(2)).Return(map[int]*substate.Substate{
		0: input0,
	}, nil)

	dst.EXPECT().PutSubstate(input0).Return(nil).Times(2)

	err := command.Execute()
	assert.Equal(t, errors.New("error"), err)
}
