package types

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash_IsEmptyReturnsTrueForZeroHash(t *testing.T) {
	var h Hash
	assert.Equal(t, true, h.IsEmpty())
}

func TestHash_IsEmptyReturnsFalseForNonZeroHash(t *testing.T) {
	h := Hash{1}
	assert.Equal(t, false, h.IsEmpty())
}

func TestHash_StringReturnsCorrectHexFormat(t *testing.T) {
	h := Hash{1, 2, 3}
	expected := "0x" + hex.EncodeToString(h[:])
	assert.Equal(t, expected, h.String())
}

func TestHash_Uint64ReturnsCorrectValue(t *testing.T) {
	h := Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42}
	assert.Equal(t, uint64(42), h.Uint64())
}

func TestHash_BigReturnsCorrectBigInt(t *testing.T) {
	h := Hash{1, 2, 3}
	expected := new(big.Int).SetBytes(h[:])
	assert.Equal(t, expected, h.Big())
}

func TestHash_BytesReturnsCorrectByteSlice(t *testing.T) {
	h := Hash{1, 2, 3}
	assert.Equal(t, h[:], h.Bytes())
}

func TestHash_CompareReturnsCorrectOrdering(t *testing.T) {
	h1 := Hash{1}
	h2 := Hash{2}
	h3 := Hash{1}

	assert.Equal(t, -1, h1.Compare(h2))
	assert.Equal(t, 1, h2.Compare(h1))
	assert.Equal(t, 0, h1.Compare(h3))
}

func TestHash_SetBytesWithSmallerInput(t *testing.T) {
	var h Hash
	input := []byte{1, 2, 3}
	h.SetBytes(input)
	expected := append(make([]byte, 29), input...)
	assert.Equal(t, expected, h.Bytes())
}

func TestHash_SetBytesWithLargerInput(t *testing.T) {
	var h Hash
	input := bytes.Repeat([]byte{1}, 40)
	h.SetBytes(input)
	assert.Equal(t, input[len(input)-32:], h.Bytes())
}

func TestHash_MarshalTextReturnsHexString(t *testing.T) {
	h := Hash{1, 2, 3}
	text, err := h.MarshalText()
	assert.Nil(t, err)
	assert.Equal(t, []byte(h.String()), text)
}

func TestHash_UnmarshalTextWithValidHex(t *testing.T) {
	var h Hash
	input := "0x0102030000000000000000000000000000000000000000000000000000000000"
	err := h.UnmarshalText([]byte(input))
	assert.Nil(t, err)
	assert.Equal(t, []byte{1, 2, 3}, h.Bytes()[:3])
}

func TestHash_UnmarshalTextWithEmptyInput(t *testing.T) {
	var h Hash
	err := h.UnmarshalText([]byte{})
	assert.Nil(t, err)
	assert.True(t, h.IsEmpty())
}

func TestHash_BytesToHashWithExactSize(t *testing.T) {
	input := bytes.Repeat([]byte{1}, 32)
	h := BytesToHash(input)
	assert.Equal(t, input, h.Bytes())
}

func TestHash_BigToHashWithSmallNumber(t *testing.T) {
	num := big.NewInt(256)
	h := BigToHash(num)
	assert.Equal(t, num.Bytes(), h.Bytes()[32-len(num.Bytes()):])
}
