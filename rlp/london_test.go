package rlp

import (
	"math/big"
	"testing"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestNewLondonRLP(t *testing.T) {
	// given
	ss := &substate.Substate{
		InputSubstate: substate.WorldState{
			types.Address{0x01}: &substate.Account{
				Nonce:   1,
				Balance: uint256.NewInt(1000),
				Storage: map[types.Hash]types.Hash{},
				Code:    []byte{0x60},
			},
		},
		OutputSubstate: substate.WorldState{
			types.Address{0x02}: &substate.Account{
				Nonce:   2,
				Balance: uint256.NewInt(2000),
				Storage: map[types.Hash]types.Hash{},
				Code:    []byte{0x61},
			},
		},
		Env: &substate.Env{
			Coinbase:    types.Address{0x03},
			Difficulty:  big.NewInt(1000),
			GasLimit:    8000000,
			Number:      100,
			Timestamp:   1633024800,
			BlockHashes: map[uint64]types.Hash{99: {0x04}},
			BaseFee:     big.NewInt(500),
		},
		Message: &substate.Message{
			Nonce:      1,
			CheckNonce: true,
			GasPrice:   big.NewInt(100),
			Gas:        21000,
			From:       types.Address{0x05},
			To:         &types.Address{0x06},
			Value:      big.NewInt(1000),
			Data:       []byte{0x07},
			AccessList: types.AccessList{},
			GasFeeCap:  big.NewInt(200),
			GasTipCap:  big.NewInt(50),
		},
		Result: &substate.Result{
			Status: 1,
			Bloom:  types.Bloom{0x08},
			Logs:   []*types.Log{},
		},
	}

	// when
	result, err := NewLondonRLP(ss)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.InputSubstate)
	assert.NotNil(t, result.OutputSubstate)
	assert.NotNil(t, result.Env)
	assert.NotNil(t, result.Message)
	assert.NotNil(t, result.Result)
}

func TestLondonRLP_ToRLP(t *testing.T) {
	// given
	london := londonRLP{
		InputSubstate:  WorldState{},
		OutputSubstate: WorldState{},
		Env: londonEnv{
			Coinbase:    types.Address{0x01},
			Difficulty:  big.NewInt(1000),
			GasLimit:    8000000,
			Number:      100,
			Timestamp:   1633024800,
			BlockHashes: [][2]types.Hash{},
			BaseFee:     nil,
		},
		Message: londonMessage{
			Nonce:      1,
			CheckNonce: true,
			GasPrice:   big.NewInt(100),
			Gas:        21000,
			From:       types.Address{0x02},
			To:         &types.Address{0x03},
			Value:      big.NewInt(1000),
			Data:       []byte{0x04},
		},
		Result: &Result{
			Status: 1,
			Bloom:  types.Bloom{0x05},
			Logs:   []*types.Log{},
		},
	}

	// when
	result := london.toRLP()

	// then
	assert.NotNil(t, result)
	assert.NotNil(t, result.Env)
	assert.NotNil(t, result.Message)
	assert.NotNil(t, result.Result)
}

func TestNewLondonEnv(t *testing.T) {
	// given
	env := &substate.Env{
		Coinbase:    types.Address{0x01},
		Difficulty:  big.NewInt(2000),
		GasLimit:    9000000,
		Number:      200,
		Timestamp:   1633024900,
		BlockHashes: map[uint64]types.Hash{199: {0x02}},
		BaseFee:     big.NewInt(750),
	}

	// when
	result := newLondonEnv(env)

	// then
	assert.Equal(t, env.Coinbase, result.Coinbase)
	assert.Equal(t, env.Difficulty, result.Difficulty)
	assert.Equal(t, env.GasLimit, result.GasLimit)
	assert.Equal(t, env.Number, result.Number)
	assert.Equal(t, env.Timestamp, result.Timestamp)
	assert.NotNil(t, result.BaseFee)
	assert.NotEmpty(t, result.BlockHashes)
}

func TestNewLondonMessage(t *testing.T) {
	// given
	to := types.Address{0x02}
	message := &substate.Message{
		Nonce:      3,
		CheckNonce: true,
		GasPrice:   big.NewInt(150),
		Gas:        30000,
		From:       types.Address{0x01},
		To:         &to,
		Value:      big.NewInt(2000),
		Data:       []byte{0x03, 0x04},
		AccessList: types.AccessList{},
		GasFeeCap:  big.NewInt(300),
		GasTipCap:  big.NewInt(100),
	}

	// when
	result, err := newLondonMessage(message)

	// then
	assert.NoError(t, err)
	assert.Equal(t, message.Nonce, result.Nonce)
	assert.Equal(t, message.CheckNonce, result.CheckNonce)
	assert.Equal(t, message.GasPrice, result.GasPrice)
	assert.Equal(t, message.Gas, result.Gas)
	assert.Equal(t, message.From, result.From)
	assert.Equal(t, message.To, result.To)
	assert.Equal(t, message.Value, result.Value)
	assert.Nil(t, result.InitCodeHash)
}

func TestNewLondonMessage_ContractCreation(t *testing.T) {
	// given
	message := &substate.Message{
		Nonce:      1,
		CheckNonce: true,
		GasPrice:   big.NewInt(100),
		Gas:        50000,
		From:       types.Address{0x01},
		To:         nil, // contract creation
		Value:      big.NewInt(0),
		Data:       []byte{0x60, 0x60, 0x60},
		AccessList: types.AccessList{},
		GasFeeCap:  big.NewInt(200),
		GasTipCap:  big.NewInt(50),
	}

	// when
	result, err := newLondonMessage(message)

	// then
	assert.NoError(t, err)
	assert.Nil(t, result.To)
	assert.NotNil(t, result.InitCodeHash)
	assert.Nil(t, result.Data)
}

func TestCreateBlockHashes(t *testing.T) {
	// given
	blockHashesMap := map[uint64]types.Hash{
		100: {0x01},
		101: {0x02},
	}

	// when
	result := createBlockHashes(blockHashesMap)

	// then
	assert.Len(t, result, 2)
	assert.NotEmpty(t, result[0][0])
	assert.NotEmpty(t, result[0][1])
}

func TestLondonEnv_ToEnv(t *testing.T) {
	// given
	baseFee := types.Hash{0x01}
	londonEnv := londonEnv{
		Coinbase:    types.Address{0x02},
		Difficulty:  big.NewInt(3000),
		GasLimit:    10000000,
		Number:      300,
		Timestamp:   1633025000,
		BlockHashes: [][2]types.Hash{},
		BaseFee:     &baseFee,
	}

	// when
	result := londonEnv.toEnv()

	// then
	assert.NotNil(t, result)
	assert.Equal(t, londonEnv.Coinbase, result.Coinbase)
	assert.Equal(t, londonEnv.Difficulty, result.Difficulty)
	assert.Equal(t, londonEnv.GasLimit, result.GasLimit)
	assert.Equal(t, londonEnv.BaseFee, result.BaseFee)
}

func TestLondonMessage_ToMessage(t *testing.T) {
	// given
	to := types.Address{0x01}
	initCodeHash := types.Hash{0x02}
	londonMsg := londonMessage{
		Nonce:        2,
		CheckNonce:   false,
		GasPrice:     big.NewInt(120),
		Gas:          25000,
		From:         types.Address{0x03},
		To:           &to,
		Value:        big.NewInt(3000),
		Data:         []byte{0x05},
		InitCodeHash: &initCodeHash,
		AccessList:   types.AccessList{},
		GasFeeCap:    big.NewInt(250),
		GasTipCap:    big.NewInt(80),
	}

	// when
	result := londonMsg.toMessage()

	// then
	assert.NotNil(t, result)
	assert.Equal(t, londonMsg.Nonce, result.Nonce)
	assert.Equal(t, londonMsg.CheckNonce, result.CheckNonce)
	assert.Equal(t, londonMsg.GasPrice, result.GasPrice)
	assert.Equal(t, londonMsg.Gas, result.Gas)
	assert.Equal(t, londonMsg.InitCodeHash, result.InitCodeHash)
}
