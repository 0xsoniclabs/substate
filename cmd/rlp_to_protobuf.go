package main

import (
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/urfave/cli/v2"
	"log"
)

type RLPtoProtobufCommand struct {
	src db.SubstateDB
	dst db.SubstateDB
	ctx *cli.Context
}

func (c *RLPtoProtobufCommand) Execute() error {

	segment, err := ParseBlockSegment(c.ctx.String(BlockSegmentFlag.Name))
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
	_, err = c.dst.SetSubstateEncoding("protobuf")
	if err != nil {
		return err
	}
	return taskPool.Execute()
}

func (c *RLPtoProtobufCommand) performSubstateUpgrade(block uint64, tx int, substate *substate.Substate, taskPool *db.SubstateTaskPool) error {
	err := c.dst.PutSubstate(substate)
	if err != nil {
		log.Printf("Failed to put substate: %v", err)
		return err
	}
	return nil
}
