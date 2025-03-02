package substate

import (
	"math/big"
	"testing"

	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
)

func TestSubstate_NewSubstate(t *testing.T) {

	preState := WorldState{}
	postState := WorldState{}
	env := &Env{}
	message := &Message{}
	result := &Result{}
	block := uint64(1)
	transaction := 1

	substate := NewSubstate(preState, postState, env, message, result, block, transaction)

	assert.Equal(t, preState, substate.InputSubstate)
	assert.Equal(t, postState, substate.OutputSubstate)
	assert.Equal(t, env, substate.Env)
	assert.Equal(t, message, substate.Message)
	assert.Equal(t, result, substate.Result)
	assert.Equal(t, block, substate.Block)
	assert.Equal(t, transaction, substate.Transaction)
}

func TestSubstate_Equal(t *testing.T) {
	preState := WorldState{}
	postState := WorldState{}
	env := &Env{}
	message := &Message{}
	result := &Result{}
	block := uint64(1)
	transaction := 1

	// self test
	substate := NewSubstate(preState, postState, env, message, result, block, transaction)
	assert.Nil(t, substate.Equal(substate))

	// nil test
	assert.NotNil(t, substate.Equal(nil))

	// preState test
	candidate := NewSubstate(
		NewWorldState().Add(types.Address{1}, 1, new(big.Int).SetUint64(1), nil),
		postState, env, message, result, block, transaction,
	)
	assert.NotNil(t, substate.Equal(candidate))

	// postState test
	candidate = NewSubstate(
		preState,
		NewWorldState().Add(types.Address{1}, 1, new(big.Int).SetUint64(1), nil),
		env, message, result, block, transaction,
	)
	assert.NotNil(t, substate.Equal(candidate))

	// env test
	candidate = NewSubstate(
		preState,
		postState,
		&Env{
			Coinbase:  types.Address{},
			GasLimit:  1,
			Number:    2,
			Timestamp: 3,
		}, message, result, block, transaction,
	)
	assert.NotNil(t, substate.Equal(candidate))

	// msg test
	candidate = NewSubstate(
		preState, postState, env, &Message{
			Nonce: 1,
			Gas:   2,
		}, result, block, transaction,
	)
	assert.NotNil(t, substate.Equal(candidate))

	// res test
	candidate = NewSubstate(
		preState, postState, env, message, &Result{
			Status:  1,
			GasUsed: 2,
		}, block, transaction,
	)
	assert.NotNil(t, substate.Equal(candidate))
}

func TestSubstate_String(t *testing.T) {
	preState := WorldState{}
	postState := WorldState{}
	env := &Env{}
	message := &Message{}
	result := &Result{}
	block := uint64(1)
	transaction := 1

	substate := NewSubstate(preState, postState, env, message, result, block, transaction)

	expected := "InputSubstate: \nOutputSubstate: \nEnv World State: Coinbase: 0x0000000000000000000000000000000000000000\nDifficulty: <nil>\nGas Limit: 0\nNumber: 0\nTimestamp: 0\nBase Fee: <nil>\nBlob Base Fee: <nil>\nBlock Hashes: \nRandom: <nil>\n\nMessage World State: Nonce: 0\nCheckNonce: false\nFrom: 0x0000000000000000000000000000000000000000\nTo: <nil>\nValue: <nil>\nData: \nData Hash: <nil>\nGas Fee Cap: <nil>\nGas Tip Cap: <nil>\n\nResult World State: Status: 0Bloom: \x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00Contract Address: 0x0000000000000000000000000000000000000000Gas Used: 0\n"
	assert.Equal(t, expected, substate.String())
}
