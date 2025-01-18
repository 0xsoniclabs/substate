package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"regexp"
	"strconv"
	"strings"
)

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

type BlockSegment struct {
	First, Last uint64
}

func NewBlockSegment(first, last uint64) *BlockSegment {
	return &BlockSegment{
		First: first,
		Last:  last,
	}
}

func ParseBlockSegment(s string) (*BlockSegment, error) {
	var err error
	// <first>: first block number
	// <last>: optional, last block number
	// <siunit>: optinal, k for 1000, M for 1000000
	re := regexp.MustCompile(`^(?P<first>[0-9][0-9_]*)((-|~)(?P<last>[0-9][0-9_]*)(?P<siunit>[kM]?))?$`)
	seg := &BlockSegment{}
	if !re.MatchString(s) {
		return nil, fmt.Errorf("invalid block segment string: %q", s)
	}
	matches := re.FindStringSubmatch(s)
	first := strings.ReplaceAll(matches[re.SubexpIndex("first")], "_", "")
	seg.First, err = strconv.ParseUint(first, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid block segment first: %s", err)
	}
	last := strings.ReplaceAll(matches[re.SubexpIndex("last")], "_", "")
	if len(last) == 0 {
		seg.Last = seg.First
	} else {
		seg.Last, err = strconv.ParseUint(last, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid block segment last: %s", err)
		}
	}
	siunit := matches[re.SubexpIndex("siunit")]
	switch siunit {
	case "k":
		seg.First = seg.First*1_000 + 1
		seg.Last = seg.Last * 1_000
	case "M":
		seg.First = seg.First*1_000_000 + 1
		seg.Last = seg.Last * 1_000_000
	}
	if seg.First > seg.Last {
		return nil, fmt.Errorf("block segment first is larger than last: %v-%v", seg.First, seg.Last)
	}
	return seg, nil
}
