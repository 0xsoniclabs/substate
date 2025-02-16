package protobuf

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/types/hash"
)

func bytesPadding(data []byte, length int) []byte {
	if len(data) >= length {
		return data
	}
	temp := make([]byte, length)
	copy(temp[length-len(data):], data)
	return temp
}

func codeHash(data []byte) *types.Hash {
	h := CodeHash(data)
	return &h
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

func TestCodeHash(t *testing.T) {
	code := []byte{0x01, 0x02, 0x03, 0x04}
	expected := hash.Keccak256Hash(code)
	actual := CodeHash(code)
	assert.Equal(t, expected, actual)

	expected = hash.Keccak256Hash(nil)
	actual = CodeHash(nil)
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
		InputAlloc: &Substate_Alloc{
			Alloc: []*Substate_AllocEntry{
				{
					Account: &Substate_Account{
						Contract: &Substate_Account_Code{
							Code: []byte{0x01, 0x02, 0x03, 0x04},
						},
					},
				},
			},
		},
		OutputAlloc: &Substate_Alloc{
			Alloc: []*Substate_AllocEntry{
				{
					Account: &Substate_Account{
						Contract: &Substate_Account_Code{
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
	output := input.HashedCopy()
	inputAllocAcc := output.GetInputAlloc().GetAlloc()[0].GetAccount().GetCodeHash()
	outputAllocAcc := output.GetOutputAlloc().GetAlloc()[0].GetAccount().GetCodeHash()
	txMessageInput := output.GetTxMessage().GetInitCodeHash()

	assert.Equal(t, inputAllocAcc, HashToBytes(codeHash([]byte{0x01, 0x02, 0x03, 0x04})))
	assert.Equal(t, outputAllocAcc, HashToBytes(codeHash([]byte{0x02, 0x03, 0x04, 0x05})))
	assert.Equal(t, txMessageInput, HashToBytes(codeHash([]byte{0x03, 0x04, 0x05, 0x06})))

	var nilSubstate *Substate
	output = nilSubstate.HashedCopy()
	assert.Nil(t, output)
}
