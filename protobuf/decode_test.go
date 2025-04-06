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
		InputAlloc: &Substate_Alloc{
			Alloc: []*Substate_AllocEntry{
				{
					Address: []byte{1},
					Account: &Substate_Account{
						Balance: big.NewInt(100).Bytes(),
						Nonce:   uint64Ptr(1),
					},
				},
			},
		},
		OutputAlloc: &Substate_Alloc{
			Alloc: []*Substate_AllocEntry{
				{
					Address: []byte{2},
					Account: &Substate_Account{
						state:   protoimpl.MessageState{},
						Nonce:   uint64Ptr(2),
						Balance: big.NewInt(200).Bytes(),
						Storage: []*Substate_Account_StorageEntry{
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
		return nil, errors.New("input decode error")
	}

	s := &Substate{
		InputAlloc: &Substate_Alloc{
			Alloc: []*Substate_AllocEntry{
				{
					Address: []byte{1},
					Account: &Substate_Account{
						Contract: &Substate_Account_CodeHash{CodeHash: []byte{1}},
					},
				},
			},
		},
	}

	result, err := s.Decode(lookup, 1, 0)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Error looking up codehash")
}

func TestDecode_OutputAllocDecodeFails(t *testing.T) {
	lookup := func(hash types.Hash) ([]byte, error) {
		return nil, errors.New("input decode error")
	}

	s := &Substate{
		InputAlloc: &Substate_Alloc{},
		OutputAlloc: &Substate_Alloc{
			Alloc: []*Substate_AllocEntry{
				{
					Address: []byte{1},
					Account: &Substate_Account{
						Contract: &Substate_Account_CodeHash{CodeHash: []byte{1}},
					},
				},
			},
		},
	}

	result, err := s.Decode(lookup, 1, 0)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Error looking up codehash")
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
}

// Helper function to create uint64 pointer
func uint64Ptr(v uint64) *uint64 {
	return &v
}
