package protobuf

import (
	"math/big"
	"testing"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
)

func TestEncode_EncodeSuccess(t *testing.T) {
	ss := &substate.Substate{
		InputSubstate:  make(substate.WorldState),
		OutputSubstate: make(substate.WorldState),
		Env: &substate.Env{
			Difficulty: big.NewInt(2),
		},
		Message: &substate.Message{
			GasPrice: big.NewInt(2),
			Value:    big.NewInt(2),
		},
		Result: &substate.Result{},
	}

	encoded, err := Encode(ss, 1, 0)
	assert.Nil(t, err)
	assert.NotNil(t, encoded)
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
			Balance: big.NewInt(100),
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
		BlobHashes: []types.Hash{{1}},
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
