package protobuf

import (
	"crypto/md5"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/holiman/uint256"

	"google.golang.org/protobuf/proto"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
)

// bytesToMD5 computes the MD5 hash of the given byte slice `data` and returns it as a hexadecimal string.
//
// Parameters:
// - data: The input byte slice to be hashed.
//
// Returns:
// - A string containing the hexadecimal representation of the MD5 hash of the input data.
func bytesToMD5(data []byte) string {
	hash := md5.Sum(data)
	hashString := hex.EncodeToString(hash[:])
	return hashString
}

func TestEncode_EncodeSuccess(t *testing.T) {
	// given
	input := &substate.Substate{
		InputSubstate: substate.WorldState{
			types.Address{0x01}: &substate.Account{
				Nonce:   1,
				Balance: uint256.NewInt(1000),
				Storage: map[types.Hash]types.Hash{
					{0x01}: {0x02},
				},
				Code: []byte{0x03},
			},
		},
		OutputSubstate: substate.WorldState{
			types.Address{0x04}: &substate.Account{
				Nonce:   1,
				Balance: uint256.NewInt(2000),
				Storage: map[types.Hash]types.Hash{
					{0xCD}: {0xAB},
				},
				Code: []byte{0x07},
			},
		},
		Env: &substate.Env{
			Coinbase:    types.Address{0x01},
			GasLimit:    1000000,
			Number:      1,
			Timestamp:   1633024800,
			BlockHashes: map[uint64]types.Hash{1: {0x02}},
			BaseFee:     big.NewInt(1000),
			BlobBaseFee: big.NewInt(2000),
			Difficulty:  big.NewInt(3000),
			Random:      &types.Hash{0x03},
		},
		Message: &substate.Message{
			Nonce:          1,
			CheckNonce:     true,
			GasPrice:       big.NewInt(100),
			Gas:            21000,
			From:           types.Address{0x04},
			To:             &types.Address{0x05},
			Value:          big.NewInt(500),
			Data:           []byte{0x06},
			ProtobufTxType: proto.Int32(0),
			AccessList: []types.AccessTuple{
				{
					Address: types.Address{0x07},
					StorageKeys: []types.Hash{
						{0x08},
					},
				},
			},
			GasFeeCap:     big.NewInt(1000),
			GasTipCap:     big.NewInt(2000),
			BlobGasFeeCap: big.NewInt(3000),
			BlobHashes:    []types.Hash{{0x09}},
		},
		Result: &substate.Result{
			Status: 1,
			Bloom:  types.Bloom{0x0A},
			Logs:   []*types.Log{{Address: types.Address{0x0B}, Topics: []types.Hash{{0x0C}}, Data: []byte{0x0D}}},
		},
		Block:       1,
		Transaction: 1,
	}

	// when
	data, err := Encode(input, 0, 0)

	// then
	assert.Nil(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, "5efc7953a42150436187e9db964d96b8", bytesToMD5(data))
}

func TestEncode_EncodeWithError(t *testing.T) {
	ss := &substate.Substate{
		InputSubstate:  make(substate.WorldState),
		OutputSubstate: make(substate.WorldState),
		Env:            &substate.Env{},
		Message: &substate.Message{
			GasPrice: big.NewInt(2),
			Value:    big.NewInt(2),
		},
		Result: &substate.Result{},
	}

	encoded, err := Encode(ss, 1, 0)
	assert.NotNil(t, err)
	assert.Nil(t, encoded)
}

func TestEncode_WorldState(t *testing.T) {
	ws := substate.WorldState{
		types.Address{1}: {
			Nonce:   1,
			Balance: uint256.NewInt(100),
			Storage: map[types.Hash]types.Hash{
				{1}: {2},
			},
			Code: []byte{1, 2, 3},
		},
	}

	alloc := toProtobufAlloc(ws)
	assert.Equal(t, 1, len(alloc.Alloc))
	assert.Equal(t, ws[types.Address{1}].Balance.Bytes(), alloc.Alloc[0].Account.Balance)
}

func TestEncode_BlockEnv(t *testing.T) {
	env := &substate.Env{
		Coinbase:    types.Address{1},
		Difficulty:  big.NewInt(100),
		GasLimit:    1000,
		Number:      10,
		Timestamp:   1234,
		BlockHashes: map[uint64]types.Hash{1: {1}},
		BaseFee:     big.NewInt(5),
	}

	encoded := toProtobufBlockEnv(env)
	assert.Equal(t, env.Coinbase.Bytes(), encoded.Coinbase)
	assert.Equal(t, env.Difficulty.Bytes(), encoded.Difficulty)
	assert.Equal(t, env.GasLimit, *encoded.GasLimit)
}

func TestEncode_TxMessage(t *testing.T) {
	txType := int32(1)
	to := types.Address{2}
	msg := &substate.Message{
		Nonce:          1,
		GasPrice:       big.NewInt(100),
		Gas:            1000,
		From:           types.Address{1},
		To:             &to,
		Value:          big.NewInt(50),
		Data:           []byte{1, 2, 3},
		ProtobufTxType: &txType,
		AccessList: types.AccessList{
			{Address: types.Address{1}, StorageKeys: []types.Hash{{2}, {3}}},
		},
		BlobHashes:            []types.Hash{{1}},
		GasFeeCap:             big.NewInt(1000),
		GasTipCap:             big.NewInt(2000),
		BlobGasFeeCap:         big.NewInt(3000),
		SetCodeAuthorizations: []types.SetCodeAuthorization{{ChainID: *big.NewInt(1), Address: types.Address{2}, Nonce: 3, V: 4, R: *big.NewInt(5), S: *big.NewInt(6)}},
	}

	encoded := toProtobufTxMessage(msg)
	assert.Equal(t, msg.Nonce, *encoded.Nonce)
	assert.Equal(t, msg.GasPrice.Bytes(), encoded.GasPrice)
	assert.Equal(t, msg.From.Bytes(), encoded.From)

}

func TestEncode_TxMessageWithInitCode(t *testing.T) {
	msg := &substate.Message{
		From: types.Address{1},
		To:   nil,
		Data: []byte{1, 2, 3},
	}

	encoded := toProtobufTxMessage(msg)
	assert.NotNil(t, encoded.GetInitCodeHash())
}

func TestEncode_AccessListEntry(t *testing.T) {
	entry := &types.AccessTuple{
		Address:     types.Address{1},
		StorageKeys: []types.Hash{{2}, {3}},
	}

	encoded := toProtobufAccessListEntry(entry)
	assert.Equal(t, entry.Address.Bytes(), encoded.Address)
	assert.Equal(t, 2, len(encoded.StorageKeys))
}

func TestEncode_Result(t *testing.T) {
	result := &substate.Result{
		Status:  1,
		GasUsed: 1000,
		Logs: []*types.Log{
			{
				Address: types.Address{1},
				Topics:  []types.Hash{{2}},
				Data:    []byte{3},
			},
		},
	}

	encoded := toProtobufResult(result)
	assert.Equal(t, result.Status, *encoded.Status)
	assert.Equal(t, result.GasUsed, *encoded.GasUsed)
	assert.Equal(t, 1, len(encoded.Logs))
}

func TestEncode_Log(t *testing.T) {
	log := &types.Log{
		Address: types.Address{1},
		Topics:  []types.Hash{{2}},
		Data:    []byte{3},
	}

	encoded := toProtobufLog(log)
	assert.Equal(t, log.Address.Bytes(), encoded.Address)
	assert.Equal(t, 1, len(encoded.Topics))
	assert.Equal(t, log.Data, encoded.Data)
}
