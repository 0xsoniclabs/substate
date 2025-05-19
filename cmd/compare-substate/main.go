package main

import (
	"log"
	"os"

	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/utils"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:   "compare-substate",
		Usage:  "compare two substate databases",
		Action: compare,
		Flags: []cli.Flag{
			&utils.WorkersFlag,
			&utils.SrcDbFlag,
			&utils.TargetDbFlag,
			&utils.BlockSegmentFlag,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// compare is the main function that compares two substate databases
func compare(ctx *cli.Context) error {
	// Open src DB
	src, err := db.NewSubstateDB(ctx.String(utils.SrcDbFlag.Name), &opt.Options{
		OpenFilesCacheCapacity: 1024,
		BlockCacheCapacity:     50 * opt.MiB,
		WriteBuffer:            25 * opt.MiB,
		ReadOnly:               true,
	}, nil, nil)
	if err != nil {
		return err
	}
	defer src.Close()

	// Open target DB
	target, err := db.NewSubstateDB(ctx.String(utils.TargetDbFlag.Name), &opt.Options{
		OpenFilesCacheCapacity: 1024,
		BlockCacheCapacity:     50 * opt.MiB,
		WriteBuffer:            25 * opt.MiB,
		ReadOnly:               true,
	}, nil, nil)
	if err != nil {
		return err
	}
	defer target.Close()

	segment, err := utils.ParseBlockSegment(ctx.String(utils.BlockSegmentFlag.Name))
	if err != nil {
		return err
	}

	return utils.Compare(ctx, src, target, ctx.Int(utils.WorkersFlag.Name), segment.First, segment.Last)
}
