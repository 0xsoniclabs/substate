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

func generateBinarySearchInList(list []int, target int) int {
	low, high := 0, len(list)-1
	for low <= high {
		mid := (low + high) / 2
		if list[mid] < target {
			low = mid + 1
		} else if list[mid] > target {
			high = mid - 1
		} else {
			return mid // found the target
		}
	}
	return -1 // target not found
}
