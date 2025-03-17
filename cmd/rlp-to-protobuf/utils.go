package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type blockSegment struct {
	First, Last uint64
}

func newBlockSegment(first, last uint64) *blockSegment {
	return &blockSegment{
		First: first,
		Last:  last,
	}
}

// parseBlockSegment parses a string representation of a block segment into a blockSegment struct.
// The string can be in the format "first" or "first-last" with optional SI units (k for 1000, M for 1000000).
// If the last block number is not provided, it is assumed to be the same as the first block number.
//
// Parameters:
// - s: The string representation of the block segment.
//
// Returns:
// - A pointer to a blockSegment struct with the parsed first and last block numbers.
// - An error if the string is not in a valid format or if the block numbers are invalid.
func parseBlockSegment(s string) (*blockSegment, error) {
	var err error
	// <first>: first block number
	// <last>: optional, last block number
	// <siunit>: optinal, k for 1000, M for 1000000
	re := regexp.MustCompile(`^(?P<first>[0-9][0-9_]*)((-|~)(?P<last>[0-9][0-9_]*)(?P<siunit>[kM]?))?$`)
	seg := &blockSegment{}
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
