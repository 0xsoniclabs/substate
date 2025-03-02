package protobuf

import (
	"math/big"
	"testing"

	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestUtils_BytesValueToHashWithNilInput(t *testing.T) {
	result := BytesValueToHash(nil)
	assert.Nil(t, result)
}

func TestUtils_BytesValueToHashWithValidInput(t *testing.T) {
	input := []byte{1, 2, 3}
	bv := &wrapperspb.BytesValue{Value: input}
	result := BytesValueToHash(bv)
	assert.Equal(t, types.BytesToHash(input), *result)
}

func TestUtils_BytesValueToBigIntWithNilInput(t *testing.T) {
	result := BytesValueToBigInt(nil)
	assert.Nil(t, result)
}

func TestUtils_BytesValueToBigIntWithValidInput(t *testing.T) {
	num := big.NewInt(256)
	bv := &wrapperspb.BytesValue{Value: num.Bytes()}
	result := BytesValueToBigInt(bv)
	assert.Equal(t, num, result)
}

func TestUtils_BytesValueToAddressWithNilInput(t *testing.T) {
	result := BytesValueToAddress(nil)
	assert.Nil(t, result)
}

func TestUtils_BytesValueToAddressWithValidInput(t *testing.T) {
	addr := types.Address{1, 2, 3}
	bv := &wrapperspb.BytesValue{Value: addr.Bytes()}
	result := BytesValueToAddress(bv)
	assert.Equal(t, addr, *result)
}

func TestUtils_AddressToWrapperspbBytesWithNilInput(t *testing.T) {
	result := AddressToWrapperspbBytes(nil)
	assert.Nil(t, result)
}

func TestUtils_AddressToWrapperspbBytesWithValidInput(t *testing.T) {
	addr := &types.Address{1, 2, 3}
	result := AddressToWrapperspbBytes(addr)
	assert.Equal(t, addr.Bytes(), result.Value)
}

func TestUtils_HashToWrapperspbBytesWithNilInput(t *testing.T) {
	result := HashToWrapperspbBytes(nil)
	assert.Nil(t, result)
}

func TestUtils_HashToWrapperspbBytesWithValidInput(t *testing.T) {
	hash := &types.Hash{1, 2, 3}
	result := HashToWrapperspbBytes(hash)
	assert.Equal(t, hash.Bytes(), result.Value)
}

func TestUtils_BigIntToWrapperspbBytesWithNilInput(t *testing.T) {
	result := BigIntToWrapperspbBytes(nil)
	assert.Nil(t, result)
}

func TestUtils_BigIntToWrapperspbBytesWithValidInput(t *testing.T) {
	num := big.NewInt(256)
	result := BigIntToWrapperspbBytes(num)
	assert.Equal(t, num.Bytes(), result.Value)
}

func TestUtils_BytesToBigIntWithNilInput(t *testing.T) {
	result := BytesToBigInt(nil)
	assert.Nil(t, result)
}

func TestUtils_BytesToBigIntWithValidInput(t *testing.T) {
	num := big.NewInt(256)
	result := BytesToBigInt(num.Bytes())
	assert.Equal(t, num, result)
}

func TestUtils_BigIntToBytesWithNilInput(t *testing.T) {
	result := BigIntToBytes(nil)
	assert.Nil(t, result)
}

func TestUtils_BigIntToBytesWithValidInput(t *testing.T) {
	num := big.NewInt(256)
	result := BigIntToBytes(num)
	assert.Equal(t, num.Bytes(), result)
}
