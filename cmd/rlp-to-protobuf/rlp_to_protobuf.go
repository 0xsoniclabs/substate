package main

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/urfave/cli/v2"
)

type rlpToProtobufCommand struct {
	src db.SubstateDB
	dst db.SubstateDB
	ctx *cli.Context
}

func (c *rlpToProtobufCommand) execute() error {

	segment, err := parseBlockSegment(c.ctx.String(BlockSegmentFlag.Name))
	if err != nil {
		return err
	}

	taskPool := &db.SubstateTaskPool{
		Name:     "rlp-to-protobuf",
		TaskFunc: c.performSubstateUpgrade,

		First: segment.First,
		Last:  segment.Last,

		Workers:         c.ctx.Int(WorkersFlag.Name),
		SkipTransferTxs: c.ctx.Bool(SkipTransferTxsFlag.Name),
		SkipCallTxs:     c.ctx.Bool(SkipCallTxsFlag.Name),
		SkipCreateTxs:   c.ctx.Bool(SkipCreateTxsFlag.Name),

		Ctx: c.ctx,

		DB: c.src,
	}
	err = c.dst.SetSubstateEncoding(db.ProtobufEncodingSchema)
	if err != nil {
		return err
	}
	return taskPool.Execute()
}

func (c *rlpToProtobufCommand) performSubstateUpgrade(
	block uint64,
	tx int,
	substate *substate.Substate,
	taskPool *db.SubstateTaskPool,
) error {
	err := c.dst.PutSubstate(substate)
	if err != nil {
		return fmt.Errorf("failed to put substate: %w", err)
	}
	return nil
}

func RunRlpToProtobuf(ctx *cli.Context) error {
	// Open old DB
	src, err := db.NewSubstateDB(ctx.String(SrcDbFlag.Name), &opt.Options{
		OpenFilesCacheCapacity: 1024,
		BlockCacheCapacity:     50 * opt.MiB,
		WriteBuffer:            25 * opt.MiB,
		ReadOnly:               true,
	}, nil, nil)
	if err != nil {
		return err
	}
	defer src.Close()

	// Open new DB
	dst, err := db.NewSubstateDB(ctx.String(DstDbFlag.Name), &opt.Options{
		OpenFilesCacheCapacity: 1024,
		BlockCacheCapacity:     50 * opt.MiB,
		WriteBuffer:            25 * opt.MiB,
		ReadOnly:               false,
	}, nil, nil)
	if err != nil {
		return err
	}
	defer dst.Close()

	command := rlpToProtobufCommand{
		src: src,
		dst: dst,
		ctx: ctx,
	}
	return command.execute()
}
