package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBlockSegment(t *testing.T) {
	// success cases
	mapInputOutput := map[string]*blockSegment{
		"1":        newBlockSegment(1, 1),
		"1k":       newBlockSegment(1000, 1000),
		"0-1M":     newBlockSegment(0, 1000000),
		"1-200":    newBlockSegment(1, 200),
		"0-1k":     newBlockSegment(0, 1000),
		"1k-1M":    newBlockSegment(1000, 1000000),
		"1M-2M":    newBlockSegment(1000000, 2000000),
		"10-100":   newBlockSegment(10, 100),
		"1k-5k":    newBlockSegment(1000, 5000),
		"5k-10000": newBlockSegment(5000, 10000),
	}

	for input, expected := range mapInputOutput {
		value, err := parseBlockSegment(input)
		assert.Nil(t, err)
		assert.Equal(t, expected, value)
	}

	// error cases
	mapInputOutputError := map[string]string{
		"1-0":                      "block segment first (1) is greater than last (0)",
		"1K":                       "invalid block segment string: \"1K\"",
		"1K-1M":                    "invalid block segment string: \"1K-1M\"",
		"1k-1G":                    "invalid block segment string: \"1k-1G\"",
		"1-100000000000000000000M": "invalid block segment last: strconv.ParseUint: parsing \"100000000000000000000\": value out of range",
		"900000000000000000000-900000000000000000000M": "invalid block segment first: strconv.ParseUint: parsing \"900000000000000000000\": value out of range",
	}
	for input, expected := range mapInputOutputError {
		value, err := parseBlockSegment(input)
		assert.Nil(t, value)
		assert.Equal(t, expected, err.Error())
	}
}
