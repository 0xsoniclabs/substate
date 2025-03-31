package main

import (
	"log"
	"os"

	"github.com/0xsoniclabs/substate/utils"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:   "rlp-to-protobuf",
		Usage:  "Convert rlp encoded substate to protobuf encoded substate",
		Action: RunRlpToProtobuf,
		Flags: []cli.Flag{
			&utils.WorkersFlag,
			&utils.SrcDbFlag,
			&utils.DstDbFlag,
			&utils.SkipTransferTxsFlag,
			&utils.SkipCallTxsFlag,
			&utils.SkipCreateTxsFlag,
			&utils.BlockSegmentFlag,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
