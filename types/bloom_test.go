package types

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBloom_BytesToBloomWithValidInput(t *testing.T) {
	input := bytes.Repeat([]byte{1}, BloomByteLength)
	bloom := BytesToBloom(input)
	assert.Equal(t, input, bloom.Bytes())
}

func TestBloom_BytesToBloomWithSmallerInput(t *testing.T) {
	input := []byte{1, 2, 3}
	bloom := BytesToBloom(input)
	expected := make([]byte, BloomByteLength)
	copy(expected[BloomByteLength-len(input):], input)
	assert.Equal(t, expected, bloom.Bytes())
}

func TestBloom_SetBytesWithValidInput(t *testing.T) {
	var bloom Bloom
	input := bytes.Repeat([]byte{1}, BloomByteLength)
	bloom.SetBytes(input)
	assert.Equal(t, input, bloom.Bytes())
}

func TestBloom_SetBytesWithSmallerInput(t *testing.T) {
	var bloom Bloom
	input := []byte{1, 2, 3}
	bloom.SetBytes(input)
	expected := make([]byte, BloomByteLength)
	copy(expected[BloomByteLength-len(input):], input)
	assert.Equal(t, expected, bloom.Bytes())
}

func TestBloom_BytesReturnsCorrectLength(t *testing.T) {
	var bloom Bloom
	assert.Equal(t, BloomByteLength, len(bloom.Bytes()))
}

func TestBloom_SetBytesPanicsOnLargeInput(t *testing.T) {
	var bloom Bloom
	input := bytes.Repeat([]byte{1}, BloomByteLength+1)
	assert.Panics(t, func() {
		bloom.SetBytes(input)
	})
}
