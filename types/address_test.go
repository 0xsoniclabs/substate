package types

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"encoding/hex"
)

func TestAddress_EmptyAddressReturnsZeroBytes(t *testing.T) {
	var addr Address
	expected := make([]byte, AddressLength)
	assert.Equal(t, expected, addr.Bytes())
}

func TestAddress_HexToAddressWithoutPrefixCreatesCorrectAddress(t *testing.T) {
	input := "1234567890abcdef1234567890abcdef12345678"
	addr := HexToAddress(input)
	assert.Equal(t, input, hex.EncodeToString(addr.Bytes()))
}

func TestAddress_HexToAddressWithPrefixCreatesCorrectAddress(t *testing.T) {
	input := "0x1234567890abcdef1234567890abcdef12345678"
	addr := HexToAddress(input)
	assert.Equal(t, input[2:], hex.EncodeToString(addr.Bytes()))
}

func TestAddress_BytesToAddressWithExactLengthCreatesCorrectAddress(t *testing.T) {
	b := bytes.Repeat([]byte{1}, AddressLength)
	addr := BytesToAddress(b)
	assert.Equal(t, b, addr.Bytes())
}

func TestAddress_BytesToAddressWithLongerInputTruncatesFromLeft(t *testing.T) {
	b := bytes.Repeat([]byte{1}, AddressLength+10)
	addr := BytesToAddress(b)
	assert.Equal(t, b[len(b)-AddressLength:], addr.Bytes())
}

func TestAddress_BytesToAddressWithShorterInputPadsWithZeros(t *testing.T) {
	b := []byte{1, 2, 3}
	addr := BytesToAddress(b)
	expected := append(make([]byte, AddressLength-len(b)), b...)
	assert.Equal(t, expected, addr.Bytes())
}

func TestAddress_SetBytesWithExactLengthSetsCorrectly(t *testing.T) {
	var addr Address
	b := bytes.Repeat([]byte{1}, AddressLength)
	addr.SetBytes(b)
	assert.Equal(t, b, addr.Bytes())
}

func TestAddress_UnmarshalTextWithEmptyInputCreatesZeroAddress(t *testing.T) {
	var addr Address
	err := addr.UnmarshalText([]byte{})
	assert.Nil(t, err)
	assert.Equal(t, make([]byte, AddressLength), addr.Bytes())
}

func TestAddress_MarshalTextAndUnmarshalTextAreReversible(t *testing.T) {
	original := HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	text, err := original.MarshalText()
	assert.Nil(t, err)

	var decoded Address
	err = decoded.UnmarshalText(text)
	assert.Nil(t, err)
	assert.Equal(t, original, decoded)
}

func TestAddress_FromHexWithOddLengthPadsWithLeadingZero(t *testing.T) {
	hex := "0x123"
	result := FromHex(hex)
	assert.Equal(t, []byte{0x01, 0x23}, result)
}
