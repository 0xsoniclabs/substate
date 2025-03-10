package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:   "rlp-to-protobuf",
		Usage:  "Convert rlp encoded substate to protobuf encoded substate",
		Action: RunRlpToProtobuf,
		Flags: []cli.Flag{
			&WorkersFlag,
			&SrcDbFlag,
			&DstDbFlag,
			&BlockSegmentFlag,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
