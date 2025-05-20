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
		Name: "compare-substate",
		Usage: "Compare two substate databases for equality. " +
			"The tool iterates trough both databases, pairs up the corresponding substates and compares them for equality.",
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
	defer func() {
		if err = src.Close(); err != nil {
			log.Printf("Error closing source DB: %v", err)
		}
	}()

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
	defer func() {
		if err = target.Close(); err != nil {
			log.Printf("Error closing target DB: %v", err)
		}
	}()

	segment, err := utils.ParseBlockSegment(ctx.String(utils.BlockSegmentFlag.Name))
	if err != nil {
		return err
	}

	return utils.Compare(ctx, src, target, ctx.Int(utils.WorkersFlag.Name), segment.First, segment.Last)
}
