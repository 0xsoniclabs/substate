package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseBlockSegment(t *testing.T) {
	// success cases
	mapInputOutput := map[string]*BlockSegment{
		"1":     NewBlockSegment(1, 1),
		"0-1M":  NewBlockSegment(1, 1000000),
		"1-200": NewBlockSegment(1, 200),
		"0-1k":  NewBlockSegment(1, 1000),
	}

	for input, expected := range mapInputOutput {
		value, err := ParseBlockSegment(input)
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
		value, err := ParseBlockSegment(input)
		assert.Nil(t, value)
		assert.Equal(t, expected, err.Error())
	}
}
