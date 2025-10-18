package utils

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

func TestKeccak256Hash(t *testing.T) {
	tests := []struct {
		name        string
		input       [][]byte
		expectError bool
	}{
		{
			name:        "empty input",
			input:       [][]byte{},
			expectError: false,
		},
		{
			name:        "single byte array",
			input:       [][]byte{[]byte("hello")},
			expectError: false,
		},
		{
			name:        "multiple byte arrays",
			input:       [][]byte{[]byte("hello"), []byte("world")},
			expectError: false,
		},
		{
			name:        "nil byte array",
			input:       [][]byte{nil},
			expectError: false,
		},
		{
			name:        "empty byte array",
			input:       [][]byte{{}},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Keccak256Hash(tt.input...)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestBytesToBloom(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expectError bool
	}{
		{
			name:        "nil input",
			input:       nil,
			expectError: false,
		},
		{
			name:        "empty byte array",
			input:       []byte{},
			expectError: false,
		},
		{
			name:        "valid 256 bytes",
			input:       make([]byte, 256),
			expectError: false,
		},
		{
			name:        "small byte array",
			input:       []byte{1, 2, 3, 4, 5},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BytesToBloom(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMust(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		err         error
		expectPanic bool
	}{
		{
			name:        "no error",
			value:       42,
			err:         nil,
			expectPanic: false,
		},
		{
			name:        "with error",
			value:       0,
			err:         assert.AnError,
			expectPanic: true,
		},
		{
			name:        "string value no error",
			value:       "test",
			err:         nil,
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					Must(tt.value, tt.err)
				})
			} else {
				assert.NotPanics(t, func() {
					result := Must(tt.value, tt.err)
					assert.Equal(t, tt.value, result)
				})
			}
		})
	}
}
