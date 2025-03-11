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
// The string can be in the following formats:
//   - Single block: "1", "10", "1k" (1000), "1M" (1000000)
//   - Block range: "1-10", "1k-5k" (1000-5000), "1k-1M" (1000-1000000), "1-1M" (1-1000000)
//
// SI units can be applied to either number independently:
//   - k: multiply by 1,000 (kilo)
//   - M: multiply by 1,000,000 (mega)
//
// Parameters:
//   - s: The string representation of the block segment.
//
// Returns:
//   - A pointer to a blockSegment struct with the parsed first and last block numbers.
//   - An error if the string is not in a valid format or if the block numbers are invalid.
func parseBlockSegment(s string) (*blockSegment, error) {
	var err error
	// <first>: first block number with optional SI unit (k for 1000, M for 1000000)
	// <last>: optional, last block number with optional SI unit
	re := regexp.MustCompile(`^(?P<first>[0-9][0-9_]*)(?P<firstunit>[kM]?)((-|~)(?P<last>[0-9][0-9_]*)(?P<lastunit>[kM]?))?$`)
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

	// Apply SI unit for first number
	firstUnit := matches[re.SubexpIndex("firstunit")]
	switch firstUnit {
	case "k":
		seg.First = seg.First * 1_000
	case "M":
		seg.First = seg.First * 1_000_000
	}

	last := strings.ReplaceAll(matches[re.SubexpIndex("last")], "_", "")
	if len(last) == 0 {
		seg.Last = seg.First
	} else {
		seg.Last, err = strconv.ParseUint(last, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid block segment last: %s", err)
		}

		// Apply SI unit for last number
		lastUnit := matches[re.SubexpIndex("lastunit")]
		switch lastUnit {
		case "k":
			seg.Last = seg.Last * 1_000
		case "M":
			seg.Last = seg.Last * 1_000_000
		}
	}

	if seg.First > seg.Last {
		return nil, fmt.Errorf("block segment first (%v) is greater than last (%v)", seg.First, seg.Last)
	}
	return seg, nil
}
