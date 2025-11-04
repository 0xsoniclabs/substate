package protobuf

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/types/hash"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// bytesPadding pads the input byte slice `data` with leading zeros to ensure it has the specified `length`.
// If the length of `data` is already greater than or equal to `length`, it returns `data` unchanged.
// Otherwise, it creates a new byte slice of the specified `length`, copies `data` into the end of the new slice,
// and returns the new slice.
//
// Parameters:
// - data: The input byte slice to be padded.
// - length: The desired length of the output byte slice.
//
// Returns:
// - A new byte slice of the specified `length` with `data` copied into the end, padded with leading zeros if necessary.
func bytesPadding(data []byte, length int) []byte {
	if len(data) >= length {
		return data
	}
	temp := make([]byte, length)
	copy(temp[length-len(data):], data)
	return temp
}

// codeHash computes the Keccak-256 hash of the given byte slice `data` and returns it as a pointer to a types.Hash.
//
// Parameters:
// - data: The input byte slice to be hashed.
//
// Returns:
// - A pointer to a types.Hash containing the Keccak-256 hash of the input data.
func codeHash(t *testing.T, data []byte) *types.Hash {
	h, err := CodeHash(data)
	assert.Nil(t, err)
	return &h
}

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

func TestBytesValueToHash(t *testing.T) {
	bytesValue := &wrapperspb.BytesValue{Value: []byte{0x01, 0x02, 0x03, 0x04}}
	expectedHash := types.BytesToHash([]byte{0x01, 0x02, 0x03, 0x04})
	hash := BytesValueToHash(bytesValue)
	assert.NotNil(t, hash)
	assert.Equal(t, expectedHash, *hash)

	hash = BytesValueToHash(nil)
	assert.Nil(t, hash)
}

func TestBytesValueToBigInt(t *testing.T) {
	bytesValue := &wrapperspb.BytesValue{Value: []byte{0x01, 0x02, 0x03, 0x04}}
	expectedBigInt := new(big.Int).SetBytes([]byte{0x01, 0x02, 0x03, 0x04})
	bigInt := BytesValueToBigInt(bytesValue)
	assert.NotNil(t, bigInt)
	assert.Equal(t, expectedBigInt, bigInt)

	bigInt = BytesValueToBigInt(nil)
	assert.Nil(t, bigInt)
}

func TestBytesValueToAddress(t *testing.T) {
	bytesValue := &wrapperspb.BytesValue{Value: []byte{0x01, 0x02, 0x03, 0x04}}
	expectedAddress := types.BytesToAddress([]byte{0x01, 0x02, 0x03, 0x04})
	address := BytesValueToAddress(bytesValue)
	assert.NotNil(t, address)
	assert.Equal(t, expectedAddress, *address)

	address = BytesValueToAddress(nil)
	assert.Nil(t, address)
}

func TestAddressToWrapperspbBytes(t *testing.T) {
	address := types.BytesToAddress([]byte{0x01, 0x02, 0x03, 0x04})
	expectedBytesValue := bytesPadding([]byte{0x01, 0x02, 0x03, 0x04}, 20)
	bytesValue := AddressToWrapperspbBytes(&address)
	assert.NotNil(t, bytesValue)
	assert.Equal(t, expectedBytesValue, bytesValue.GetValue())

	bytesValue = AddressToWrapperspbBytes(nil)
	assert.Nil(t, bytesValue)
}

func TestHashToWrapperspbBytes(t *testing.T) {
	hash := types.BytesToHash([]byte{0x01, 0x02, 0x03, 0x04})
	expectedBytesValue := bytesPadding([]byte{0x01, 0x02, 0x03, 0x04}, 32)
	bytesValue := HashToWrapperspbBytes(&hash)
	assert.NotNil(t, bytesValue)
	assert.Equal(t, expectedBytesValue, bytesValue.GetValue())

	bytesValue = HashToWrapperspbBytes(nil)
	assert.Nil(t, bytesValue)
}

func TestBigIntToWrapperspbBytes(t *testing.T) {
	bigInt := new(big.Int).SetBytes([]byte{0x01, 0x02, 0x03, 0x04})
	expectedBytesValue := &wrapperspb.BytesValue{Value: []byte{0x01, 0x02, 0x03, 0x04}}
	bytesValue := BigIntToWrapperspbBytes(bigInt)
	assert.NotNil(t, bytesValue)
	assert.True(t, bytes.Equal(bytesValue.GetValue(), expectedBytesValue.GetValue()))

	bytesValue = BigIntToWrapperspbBytes(nil)
	assert.Nil(t, bytesValue)
}

func TestBytesToBigInt(t *testing.T) {
	bytes := []byte{0x01, 0x02, 0x03, 0x04}
	expectedBigInt := new(big.Int).SetBytes(bytes)
	bigInt := BytesToBigInt(bytes)
	assert.NotNil(t, bigInt)
	assert.Equal(t, expectedBigInt, bigInt)

	bigInt = BytesToBigInt(nil)
	assert.Nil(t, bigInt)
}

func TestBytesToUint256(t *testing.T) {
	b := []byte{0x01, 0x02, 0x03, 0x04}

	actual := BytesToUint256(b)
	expected := new(uint256.Int).SetBytes(b)
	assert.Equal(t, expected, actual)
}

func TestCodeHash(t *testing.T) {
	code := []byte{0x01, 0x02, 0x03, 0x04}
	expected := hash.Keccak256Hash(code)
	actual, err := CodeHash(code)
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)

	expected = hash.Keccak256Hash(nil)
	actual, err = CodeHash(nil)
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func TestHashToBytes(t *testing.T) {
	hash := types.BytesToHash([]byte{0x01, 0x02, 0x03, 0x04})
	expectedBytes := bytesPadding([]byte{0x1, 0x2, 0x3, 0x4}, 32)
	actualBytes := HashToBytes(&hash)
	assert.Equal(t, expectedBytes, actualBytes)

	actualBytes = HashToBytes(nil)
	assert.Nil(t, actualBytes)
}

func TestHashedCopy(t *testing.T) {
	input := &Substate{
		InputAlloc: &Alloc{
			Alloc: []*AllocEntry{
				{
					Account: &Account{
						Contract: &Account_Code{
							Code: []byte{0x01, 0x02, 0x03, 0x04},
						},
					},
				},
			},
		},
		OutputAlloc: &Alloc{
			Alloc: []*AllocEntry{
				{
					Account: &Account{
						Contract: &Account_Code{
							Code: []byte{0x02, 0x03, 0x04, 0x05},
						},
					},
				},
			},
		},
		TxMessage: &Substate_TxMessage{
			Input: &Substate_TxMessage_Data{
				Data: []byte{0x03, 0x04, 0x05, 0x06},
			},
		},
	}
	output, err := input.HashedCopy()
	assert.Nil(t, err)
	inputAllocAcc := output.GetInputAlloc().GetAlloc()[0].GetAccount().GetCodeHash()
	outputAllocAcc := output.GetOutputAlloc().GetAlloc()[0].GetAccount().GetCodeHash()
	txMessageInput := output.GetTxMessage().GetInitCodeHash()

	assert.Equal(t, inputAllocAcc, HashToBytes(codeHash(t, []byte{0x01, 0x02, 0x03, 0x04})))
	assert.Equal(t, outputAllocAcc, HashToBytes(codeHash(t, []byte{0x02, 0x03, 0x04, 0x05})))
	assert.Equal(t, txMessageInput, HashToBytes(codeHash(t, []byte{0x03, 0x04, 0x05, 0x06})))

	var nilSubstate *Substate
	output, err = nilSubstate.HashedCopy()
	assert.NoError(t, err)
	assert.Nil(t, output)
}
