package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBlockSegment(t *testing.T) {
	// success cases
	mapInputOutput := map[string]*blockSegment{
		"1":     newBlockSegment(1, 1),
		"0-1M":  newBlockSegment(1, 1000000),
		"1-200": newBlockSegment(1, 200),
		"0-1k":  newBlockSegment(1, 1000),
	}

	for input, expected := range mapInputOutput {
		value, err := parseBlockSegment(input)
		assert.Nil(t, err)
		assert.Equal(t, expected, value)
	}

	// error cases
	mapInputOutputError := map[string]string{
		"1-0":                      "block segment first is larger than last: 1-0",
		"1M":                       "invalid block segment string: \"1M\"",
		"1-1M":                     "block segment first is larger than last: 1000001-1000000",
		"1-100000000000000000000M": "invalid block segment last: strconv.ParseUint: parsing \"100000000000000000000\": value out of range",
		"900000000000000000000-900000000000000000000M": "invalid block segment first: strconv.ParseUint: parsing \"900000000000000000000\": value out of range",
	}
	for input, expected := range mapInputOutputError {
		value, err := parseBlockSegment(input)
		assert.Nil(t, value)
		assert.Equal(t, expected, err.Error())
	}
}
