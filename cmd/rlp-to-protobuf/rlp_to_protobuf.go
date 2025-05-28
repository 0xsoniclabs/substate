package main

import (
	"errors"
	"fmt"

	"github.com/0xsoniclabs/substate/utils"
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
	segment, err := utils.ParseBlockSegment(c.ctx.String(utils.BlockSegmentFlag.Name))
	if err != nil {
		return err
	}

	taskPool := &db.SubstateTaskPool{
		Name:     "rlp-to-protobuf",
		TaskFunc: c.performSubstateUpgrade,

		First: segment.First,
		Last:  segment.Last,

		Workers:         c.ctx.Int(utils.WorkersFlag.Name),
		SkipTransferTxs: c.ctx.Bool(utils.SkipTransferTxsFlag.Name),
		SkipCallTxs:     c.ctx.Bool(utils.SkipCallTxsFlag.Name),
		SkipCreateTxs:   c.ctx.Bool(utils.SkipCreateTxsFlag.Name),

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

func RunRlpToProtobuf(ctx *cli.Context) (outErr error) {
	// Open old DB
	src, err := db.NewSubstateDB(ctx.String(utils.SrcDbFlag.Name), &opt.Options{
		OpenFilesCacheCapacity: 1024,
		BlockCacheCapacity:     50 * opt.MiB,
		WriteBuffer:            25 * opt.MiB,
		ReadOnly:               true,
	}, nil, nil)
	if err != nil {
		return err
	}
	defer func() {
		e := src.Close()
		if e != nil {
			outErr = errors.Join(outErr, e)
		}
	}()

	// Open new DB
	dst, err := db.NewSubstateDB(ctx.String(utils.DstDbFlag.Name), &opt.Options{
		OpenFilesCacheCapacity: 1024,
		BlockCacheCapacity:     50 * opt.MiB,
		WriteBuffer:            25 * opt.MiB,
		ReadOnly:               false,
	}, nil, nil)
	if err != nil {
		return err
	}
	defer func() {
		e := dst.Close()
		if e != nil {
			outErr = errors.Join(outErr, e)
		}
	}()

	command := rlpToProtobufCommand{
		src: src,
		dst: dst,
		ctx: ctx,
	}
	return command.execute()
}
