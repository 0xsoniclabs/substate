package main

import (
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func RunRlpToProtobuf(ctx *cli.Context) error {
	// Open old DB
	src, err := db.NewCustomSubstateDB(ctx.String(SrcDbFlag.Name), 1024, 100, true)
	if err != nil {
		return err
	}
	defer src.Close()

	// Open new DB
	dst, err := db.NewCustomSubstateDB(ctx.String(DstDbFlag.Name), 1024, 100, false)
	if err != nil {
		return err
	}
	defer dst.Close()

	command := RLPtoProtobufCommand{
		src: src,
		dst: dst,
		ctx: ctx,
	}
	return command.Execute()
}

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
