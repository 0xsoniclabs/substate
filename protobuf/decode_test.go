package protobuf

import (
	"errors"
	"math/big"
	"testing"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestDecode_SuccessfulDecode(t *testing.T) {
	lookup := func(hash types.Hash) ([]byte, error) {
		return []byte{1, 2, 3}, nil
	}

	s := &Substate{
		InputAlloc: &Alloc{
			Alloc: []*AllocEntry{
				{
					Address: []byte{1},
					Account: &Account{
						Balance: big.NewInt(100).Bytes(),
						Nonce:   uint64Ptr(1),
					},
				},
			},
		},
		OutputAlloc: &Alloc{
			Alloc: []*AllocEntry{
				{
					Address: []byte{2},
					Account: &Account{
						state:   protoimpl.MessageState{},
						Nonce:   uint64Ptr(2),
						Balance: big.NewInt(200).Bytes(),
						Storage: []*Account_StorageEntry{
							{
								Key:   []byte{1},
								Value: []byte{1},
							},
						},
					},
				},
			},
		},
		BlockEnv: &Substate_BlockEnv{
			Coinbase:   []byte{1},
			Difficulty: big.NewInt(100).Bytes(),
			GasLimit:   uint64Ptr(1000),
			BlockHashes: []*Substate_BlockEnv_BlockHashEntry{
				{
					Key:   uint64Ptr(1),
					Value: []byte{1},
				},
			},
		},
		TxMessage: &Substate_TxMessage{
			From:     []byte{1},
			GasPrice: big.NewInt(100).Bytes(),
			Value:    big.NewInt(100).Bytes(),
			TxType:   Substate_TxMessage_TXTYPE_LEGACY.Enum(),
			AccessList: []*Substate_TxMessage_AccessListEntry{
				{
					Address:     []byte{1},
					StorageKeys: [][]byte{{1}},
				},
			},
		},
		Result: &Substate_Result{
			Status:  uint64Ptr(1),
			GasUsed: uint64Ptr(1000),
			Logs: []*Substate_Result_Log{
				{
					Address: []byte{1},
					Topics:  [][]byte{{1}},
				},
			},
		},
	}

	result, err := s.Decode(lookup, 1, 0)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(1), result.Block)
	assert.Equal(t, 0, result.Transaction)
	assert.NotNil(t, result.InputSubstate)
	assert.NotNil(t, result.OutputSubstate)
	assert.NotNil(t, result.Env)
	assert.NotNil(t, result.Message)
	assert.NotNil(t, result.Result)
}

func TestDecode_InputAllocDecodeFails(t *testing.T) {
	lookup := func(hash types.Hash) ([]byte, error) {
		return nil, errors.New("input Decode error")
	}

	s := &Substate{
		InputAlloc: &Alloc{
			Alloc: []*AllocEntry{
				{
					Address: []byte{1},
					Account: &Account{
						Contract: &Account_CodeHash{CodeHash: []byte{1}},
					},
				},
			},
		},
	}

	result, err := s.Decode(lookup, 1, 0)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error looking up codehash")
}

func TestDecode_OutputAllocDecodeFails(t *testing.T) {
	lookup := func(hash types.Hash) ([]byte, error) {
		return nil, errors.New("input Decode error")
	}

	s := &Substate{
		InputAlloc: &Alloc{},
		OutputAlloc: &Alloc{
			Alloc: []*AllocEntry{
				{
					Address: []byte{1},
					Account: &Account{
						Contract: &Account_CodeHash{CodeHash: []byte{1}},
					},
				},
			},
		},
	}

	result, err := s.Decode(lookup, 1, 0)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error looking up codehash")
}

func TestDecode_TxMessageLegacy(t *testing.T) {
	lookup := func(hash types.Hash) ([]byte, error) {
		return []byte{1, 2, 3}, nil
	}

	msg := &Substate_TxMessage{
		Nonce:    uint64Ptr(1),
		GasPrice: big.NewInt(100).Bytes(),
		Gas:      uint64Ptr(21000),
		From:     []byte{1},
		To:       &wrapperspb.BytesValue{Value: []byte{2}},
		Value:    big.NewInt(1000).Bytes(),
		Input: &Substate_TxMessage_Data{
			Data: []byte{4, 5, 6},
		},
		TxType: Substate_TxMessage_TXTYPE_LEGACY.Enum(),
	}

	result, err := msg.decode(lookup)

	expectedFrom := types.Address{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}
	expectedTo := types.Address{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2}
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.Nonce)
	assert.Equal(t, big.NewInt(100), result.GasPrice)
	assert.Equal(t, uint64(21000), result.Gas)
	assert.Equal(t, expectedFrom, result.From)
	assert.Equal(t, expectedTo, *result.To)
	assert.Equal(t, big.NewInt(1000), result.Value)
	assert.Equal(t, []byte{4, 5, 6}, result.Data)
	assert.Equal(t, int32(substate.LegacyTxType), *result.ProtobufTxType)
}

func TestDecode_TxMessageContractCreation(t *testing.T) {
	lookup := func(hash types.Hash) ([]byte, error) {
		return []byte{7, 8, 9}, nil
	}

	msg := &Substate_TxMessage{
		From:  []byte{1},
		Value: big.NewInt(0).Bytes(),
		Input: &Substate_TxMessage_InitCodeHash{
			InitCodeHash: []byte{1, 2, 3},
		},

		TxType: Substate_TxMessage_TXTYPE_LEGACY.Enum(),
	}

	result, err := msg.decode(lookup)

	assert.NoError(t, err)
	assert.Nil(t, result.To)
	assert.Equal(t, []byte{7, 8, 9}, result.Data)
}

func TestDecode_TxMessageAccessList(t *testing.T) {
	lookup := func(hash types.Hash) ([]byte, error) {
		return []byte{}, nil
	}

	msg := &Substate_TxMessage{
		From:   []byte{1},
		TxType: Substate_TxMessage_TXTYPE_ACCESSLIST.Enum(),
		AccessList: []*Substate_TxMessage_AccessListEntry{
			{
				Address:     []byte{2},
				StorageKeys: [][]byte{{3}},
			},
		},
	}

	result, err := msg.decode(lookup)

	expectedAddress := types.Address{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2}
	expectedStorageKey := types.Hash{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3}
	assert.NoError(t, err)
	assert.Len(t, result.AccessList, 1)
	assert.Equal(t, expectedAddress, result.AccessList[0].Address)
	assert.Equal(t, expectedStorageKey, result.AccessList[0].StorageKeys[0])
}

func TestDecode_TxMessageDynamicFee(t *testing.T) {
	lookup := func(hash types.Hash) ([]byte, error) {
		return []byte{}, nil
	}

	msg := &Substate_TxMessage{
		From:      []byte{1},
		TxType:    Substate_TxMessage_TXTYPE_DYNAMICFEE.Enum(),
		GasFeeCap: wrapperspb.Bytes(big.NewInt(200).Bytes()),
		GasTipCap: wrapperspb.Bytes(big.NewInt(100).Bytes()),
	}

	result, err := msg.decode(lookup)

	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(200), result.GasFeeCap)
	assert.Equal(t, big.NewInt(100), result.GasTipCap)
}

func TestDecode_TxMessageBlob(t *testing.T) {
	lookup := func(hash types.Hash) ([]byte, error) {
		return []byte{}, nil
	}

	msg := &Substate_TxMessage{
		From:          []byte{1},
		TxType:        Substate_TxMessage_TXTYPE_BLOB.Enum(),
		BlobGasFeeCap: wrapperspb.Bytes(big.NewInt(300).Bytes()),
		BlobHashes:    [][]byte{{1}, {2}},
	}

	result, err := msg.decode(lookup)

	expectedBlob1 := types.Hash{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}
	expectedBlob2 := types.Hash{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2}
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(300), result.BlobGasFeeCap)
	assert.Len(t, result.BlobHashes, 2)
	assert.Equal(t, expectedBlob1, result.BlobHashes[0])
	assert.Equal(t, expectedBlob2, result.BlobHashes[1])
}

func TestDecode_TxMessageSetCodeSuccess(t *testing.T) {
	lookup := func(hash types.Hash) ([]byte, error) {
		return []byte{1, 2, 3}, nil
	}

	msg := &Substate_TxMessage{
		Nonce:    uint64Ptr(1),
		GasPrice: big.NewInt(100).Bytes(),
		Gas:      uint64Ptr(21000),
		From:     []byte{1},
		To:       &wrapperspb.BytesValue{Value: []byte{2}},
		Value:    big.NewInt(1000).Bytes(),
		Input: &Substate_TxMessage_Data{
			Data: []byte{4, 5, 6},
		},
		SetCodeAuthorizations: []*Substate_TxMessage_SetCodeAuthorization{
			{
				ChainId: big.NewInt(1).Bytes(),
				Address: types.Address{2}.Bytes(),
				Nonce:   uint64Ptr(3),
				V:       []byte{4},
				R:       big.NewInt(5).Bytes(),
				S:       big.NewInt(6).Bytes(),
			},
		},
		TxType: Substate_TxMessage_TXTYPE_SETCODE.Enum(),
	}

	result, err := msg.decode(lookup)

	expectedFrom := types.Address{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}
	expectedTo := types.Address{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2}
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.Nonce)
	assert.Equal(t, big.NewInt(100), result.GasPrice)
	assert.Equal(t, uint64(21000), result.Gas)
	assert.Equal(t, expectedFrom, result.From)
	assert.Equal(t, expectedTo, *result.To)
	assert.Equal(t, big.NewInt(1000), result.Value)
	assert.Equal(t, []byte{4, 5, 6}, result.Data)
	assert.Equal(t, uint64(1), result.SetCodeAuthorizations[0].ChainID.Uint64())
	assert.Equal(t, types.Address{2}, result.SetCodeAuthorizations[0].Address)
	assert.Equal(t, uint64(3), result.SetCodeAuthorizations[0].Nonce)
	assert.Equal(t, uint8(4), result.SetCodeAuthorizations[0].V)
	assert.Equal(t, uint64(5), result.SetCodeAuthorizations[0].R.Uint64())
	assert.Equal(t, uint64(6), result.SetCodeAuthorizations[0].S.Uint64())
	assert.Equal(t, int32(substate.SetCodeTxType), *result.ProtobufTxType)
}

func TestDecode_TxMessageUnknownType(t *testing.T) {
	lookup := func(hash types.Hash) ([]byte, error) {
		return []byte{7, 8, 9}, nil
	}

	unknownType := Substate_TxMessage_TxType(999)
	msg := &Substate_TxMessage{
		From:   []byte{1},
		TxType: &unknownType,
	}

	result, err := msg.decode(lookup)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unknown tx type")
}

func TestDecode_TxMessageLookupError(t *testing.T) {
	lookup := func(hash types.Hash) ([]byte, error) {
		return nil, errors.New("lookup failed")
	}

	msg := &Substate_TxMessage{
		From: []byte{1},
		Input: &Substate_TxMessage_InitCodeHash{
			InitCodeHash: []byte{1},
		},
	}

	result, err := msg.decode(lookup)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to Decode tx message")
}

func TestDecode_TxMessageLookupNotFound(t *testing.T) {
	lookup := func(hash types.Hash) ([]byte, error) {
		return nil, leveldb.ErrNotFound
	}

	msg := &Substate_TxMessage{
		From:   []byte{1},
		TxType: Substate_TxMessage_TXTYPE_LEGACY.Enum(),
		Input: &Substate_TxMessage_InitCodeHash{
			InitCodeHash: []byte{1},
		},
	}

	result, err := msg.decode(lookup)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.Data)
}

func TestDecode_GetContractAddressWithTo(t *testing.T) {
	msg := &Substate_TxMessage{
		From: []byte{1},
		To:   &wrapperspb.BytesValue{Value: []byte{2}},
	}

	address := msg.getContractAddress()

	assert.Equal(t, types.Address{}, address)
}

func TestDecode_GetContractAddress(t *testing.T) {
	msg := &Substate_TxMessage{
		From:  []byte{1},
		Nonce: uint64Ptr(9),
	}

	address := msg.getContractAddress()

	assert.Equal(t, types.Address{0x94, 0xed, 0xc3, 0x20, 0x46, 0x6d, 0x68, 0xc0, 0xe8, 0xc, 0x3e, 0x6f, 0x45, 0x43, 0x75, 0xfb, 0x95, 0x7e, 0x10, 0x38}, address)

	msg = &Substate_TxMessage{
		From:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Nonce: nil,
	}

	address = msg.getContractAddress()

	assert.Equal(t, types.Address{0x43, 0x33, 0xeb, 0x62, 0x7, 0x26, 0xc2, 0xfa, 0x9, 0xf4, 0x4, 0xe5, 0x42, 0x4c, 0x6b, 0x54, 0x85, 0x2, 0x3, 0xbf}, address)
}

func TestDecode_createAddress2(t *testing.T) {
	addr := types.Address{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	nonce := uint64(18446744073709551615)
	assert.Equal(t, types.Address{0x9a, 0x53, 0x6c, 0xde, 0xb1, 0x3e, 0xc1, 0xa0, 0xc9, 0x3e, 0xcc, 0x66, 0xc9, 0xe2, 0x53, 0x47, 0x9d, 0x85, 0xde, 0x29}, createAddress(addr, nonce))

	addr = types.Address{1, 2, 3, 4}
	nonce = uint64(0)
	assert.Equal(t, types.Address{0xc7, 0x68, 0x3a, 0x8a, 0x54, 0xb2, 0x1d, 0x7e, 0x5b, 0x96, 0x24, 0xa8, 0xaf, 0xe6, 0x1c, 0x2d, 0x9b, 0x25, 0x96, 0x7c}, createAddress(addr, nonce))

	addr = types.Address{1, 2, 3, 4}
	nonce = uint64(1)
	assert.Equal(t, types.Address{0x29, 0xdb, 0x64, 0xcb, 0xeb, 0x87, 0xc2, 0xde, 0x17, 0x4, 0x5b, 0x6f, 0x67, 0xe4, 0x1e, 0x88, 0xab, 0xba, 0x6b, 0x1}, createAddress(addr, nonce))

	addr = types.Address{1, 2, 3, 4}
	nonce = uint64(65535)
	assert.Equal(t, types.Address{0x3d, 0x92, 0xa2, 0xfc, 0xc7, 0x37, 0xac, 0x3c, 0xce, 0xd4, 0x1, 0x1b, 0x81, 0x38, 0xf, 0xcc, 0xc5, 0x1e, 0x21, 0x7a}, createAddress(addr, nonce))

	addr = types.Address{1, 2, 3, 4}
	nonce = uint64(32767)
	assert.Equal(t, types.Address{0x5e, 0xb5, 0xeb, 0x5e, 0xf, 0x41, 0xd8, 0xdd, 0x9a, 0x4f, 0x24, 0xf3, 0x5d, 0x36, 0xed, 0x6b, 0x8f, 0x5d, 0xd1, 0xa9}, createAddress(addr, nonce))
}

func TestDecode_putInt(t *testing.T) {
	testCases := []struct {
		name          string
		input         uint64
		expectedSize  int
		expectedBytes []byte
	}{
		{
			name:          "single byte (< 1<<8)",
			input:         0x7F,
			expectedSize:  1,
			expectedBytes: []byte{0x7F},
		},
		{
			name:          "two bytes (< 1<<16)",
			input:         0x1234,
			expectedSize:  2,
			expectedBytes: []byte{0x12, 0x34},
		},
		{
			name:          "three bytes (< 1<<24)",
			input:         0x123456,
			expectedSize:  3,
			expectedBytes: []byte{0x12, 0x34, 0x56},
		},
		{
			name:          "four bytes (< 1<<32)",
			input:         0x12345678,
			expectedSize:  4,
			expectedBytes: []byte{0x12, 0x34, 0x56, 0x78},
		},
		{
			name:          "five bytes (< 1<<40)",
			input:         0x1234567890,
			expectedSize:  5,
			expectedBytes: []byte{0x12, 0x34, 0x56, 0x78, 0x90},
		},
		{
			name:          "six bytes (< 1<<48)",
			input:         0x123456789012,
			expectedSize:  6,
			expectedBytes: []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12},
		},
		{
			name:          "seven bytes (< 1<<56)",
			input:         0x12345678901234,
			expectedSize:  7,
			expectedBytes: []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34},
		},
		{
			name:          "eight bytes (>= 1<<56)",
			input:         0x1234567890123456,
			expectedSize:  8,
			expectedBytes: []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56},
		},
		{
			name:          "max uint64",
			input:         0xFFFFFFFFFFFFFFFF,
			expectedSize:  8,
			expectedBytes: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
	}

	for _, tc := range testCases {
		// Create a buffer large enough for any size
		buf := make([]byte, 8)

		// Call the function
		size := putInt(buf, tc.input)

		// Check the returned size
		assert.Equal(t, tc.expectedSize, size)

		// Check the bytes written to the buffer
		assert.Equal(t, tc.expectedBytes, buf[:size])
	}
}

// Helper function to create uint64 pointer
func uint64Ptr(v uint64) *uint64 {
	return &v
}
