package main

import "github.com/urfave/cli/v2"

var (
	WorkersFlag = cli.IntFlag{
		Name:    "workers",
		Aliases: []string{"w"},
		Usage:   "determines number of workers",
		Value:   4,
	}
	SrcDbFlag = cli.PathFlag{
		Name:     "src",
		Usage:    "Source Aida DB",
		Required: true,
	}
	DstDbFlag = cli.PathFlag{
		Name:     "dst",
		Usage:    "Destination Aida DB",
		Required: true,
	}
	SkipTransferTxsFlag = cli.BoolFlag{
		Name:  "skip-transfer-txs",
		Usage: "Skip executing transactions that only transfer ETH",
	}
	SkipCallTxsFlag = cli.BoolFlag{
		Name:  "skip-call-txs",
		Usage: "Skip executing CALL transactions to accounts with contract bytecode",
	}
	SkipCreateTxsFlag = cli.BoolFlag{
		Name:  "skip-create-txs",
		Usage: "Skip executing CREATE transactions",
	}
	BlockSegmentFlag = cli.StringFlag{
		Name:     "block-segment",
		Usage:    "Single block segment (e.g. 1001, 1_001, 1_001-2_000, 1-2k, 1-2M)",
		Required: true,
	}
)
