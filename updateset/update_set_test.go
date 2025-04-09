package updateset

import (
	"testing"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
)

func TestUpdateSet_ToWorldStateRLP(t *testing.T) {
	// Create test accounts
	acc1 := substate.NewAccount(1, new(uint256.Int).SetUint64(100), []byte{1, 2, 3})
	acc1.Storage = map[types.Hash]types.Hash{
		{1}: {2},
	}

	acc2 := substate.NewAccount(2, new(uint256.Int).SetUint64(200), []byte{4, 5, 6})

	// Create world state
	ws := substate.WorldState{
		types.Address{1}: acc1,
		types.Address{2}: acc2,
	}

	// Create update set
	updateSet := NewUpdateSet(ws, 10)

	// Convert to RLP
	rlpState := updateSet.ToWorldStateRLP()

	// Verify addresses and accounts length match
	assert.Equal(t, len(rlpState.Addresses), len(rlpState.Accounts))
	assert.Equal(t, 2, len(rlpState.Addresses))

	// Find the accounts in the RLP representation
	var foundAcc1, foundAcc2 bool
	for i, addr := range rlpState.Addresses {
		acc := rlpState.Accounts[i]

		if addr == (types.Address{1}) {
			foundAcc1 = true
			assert.Equal(t, uint64(1), acc.Nonce)
			assert.Equal(t, uint64(100), acc.Balance.Uint64())
			assert.Equal(t, acc1.CodeHash(), acc.CodeHash)
			assert.Equal(t, 1, len(acc.Storage))
		}

		if addr == (types.Address{2}) {
			foundAcc2 = true
			assert.Equal(t, uint64(2), acc.Nonce)
			assert.Equal(t, uint64(200), acc.Balance.Uint64())
			assert.Equal(t, acc2.CodeHash(), acc.CodeHash)
			assert.Equal(t, 0, len(acc.Storage))
		}
	}

	assert.True(t, foundAcc1)
	assert.True(t, foundAcc2)
}

func TestUpdateSet_Equal(t *testing.T) {
	// Create test accounts
	acc1 := substate.NewAccount(1, new(uint256.Int).SetUint64(100), []byte{1, 2, 3})
	acc2 := substate.NewAccount(2, new(uint256.Int).SetUint64(200), []byte{4, 5, 6})

	// Create first world state
	ws1 := substate.WorldState{
		types.Address{1}: acc1,
		types.Address{2}: acc2,
	}

	// Create identical second world state
	ws2 := substate.WorldState{
		types.Address{1}: acc1.Copy(),
		types.Address{2}: acc2.Copy(),
	}

	// Different world state
	ws3 := substate.WorldState{
		types.Address{1}: acc1.Copy(),
	}

	// Create update sets
	updateSet1 := NewUpdateSet(ws1, 10)
	updateSet2 := NewUpdateSet(ws2, 10)
	updateSet3 := NewUpdateSet(ws1, 11) // Different block number
	updateSet4 := NewUpdateSet(ws3, 10) // Different accounts

	// Set deleted accounts
	updateSet1.DeletedAccounts = []types.Address{{3}}
	updateSet2.DeletedAccounts = []types.Address{{3}}
	updateSet3.DeletedAccounts = []types.Address{{3}}
	updateSet4.DeletedAccounts = []types.Address{{3}}

	// Test equality
	assert.True(t, updateSet1.Equal(updateSet2))
	assert.False(t, updateSet1.Equal(updateSet3))
	assert.False(t, updateSet1.Equal(updateSet4))

	// Test self equality
	assert.True(t, updateSet1.Equal(updateSet1))

	// Test with different deleted accounts
	updateSet5 := NewUpdateSet(ws1, 10)
	updateSet5.DeletedAccounts = []types.Address{{4}}
	assert.False(t, updateSet1.Equal(updateSet5))
}

func TestUpdateSetRLP_ToWorldStateSuccess(t *testing.T) {
	// Create test accounts
	acc1 := substate.NewAccount(1, new(uint256.Int).SetUint64(100), []byte{1, 2, 3})
	acc1.Storage = map[types.Hash]types.Hash{
		{1}: {2},
	}

	acc2 := substate.NewAccount(2, new(uint256.Int).SetUint64(200), []byte{4, 5, 6})

	// Create world state
	ws := substate.WorldState{
		types.Address{1}: acc1,
		types.Address{2}: acc2,
	}

	// Create update set
	updateSet := NewUpdateSet(ws, 10)
	updateSet.DeletedAccounts = []types.Address{{3}}

	// Create RLP version
	rlpUpdateSet := NewUpdateSetRLP(updateSet, updateSet.DeletedAccounts)

	// Mock getCodeFunc
	getCodeFunc := func(codeHash types.Hash) ([]byte, error) {
		if codeHash == acc1.CodeHash() {
			return []byte{1, 2, 3}, nil
		}
		if codeHash == acc2.CodeHash() {
			return []byte{4, 5, 6}, nil
		}
		return nil, nil
	}

	// Convert back to world state
	newUpdateSet, err := rlpUpdateSet.ToWorldState(getCodeFunc, 10)
	assert.NoError(t, err)

	// Verify the conversion
	assert.Equal(t, uint64(10), newUpdateSet.Block)
	assert.Equal(t, 2, len(newUpdateSet.WorldState))

	// Check accounts are converted correctly
	for addr, account := range newUpdateSet.WorldState {
		if addr == (types.Address{1}) {
			assert.Equal(t, uint64(1), account.Nonce)
			assert.Equal(t, uint64(100), account.Balance.Uint64())
			assert.Equal(t, []byte{1, 2, 3}, account.Code)
			assert.Equal(t, 1, len(account.Storage))
			assert.Equal(t, types.Hash{2}, account.Storage[types.Hash{1}])
		}

		if addr == (types.Address{2}) {
			assert.Equal(t, uint64(2), account.Nonce)
			assert.Equal(t, uint64(200), account.Balance.Uint64())
			assert.Equal(t, []byte{4, 5, 6}, account.Code)
			assert.Equal(t, 0, len(account.Storage))
		}
	}
}

func TestUpdateSetRLP_ToWorldStateError(t *testing.T) {
	// Create test accounts
	acc1 := substate.NewAccount(1, new(uint256.Int).SetUint64(100), []byte{1, 2, 3})
	acc1.Storage = map[types.Hash]types.Hash{
		{1}: {2},
	}

	acc2 := substate.NewAccount(2, new(uint256.Int).SetUint64(200), []byte{4, 5, 6})

	// Create world state
	ws := substate.WorldState{
		types.Address{1}: acc1,
		types.Address{2}: acc2,
	}

	// Create update set
	updateSet := NewUpdateSet(ws, 10)
	updateSet.DeletedAccounts = []types.Address{{3}}

	// Create RLP version
	rlpUpdateSet := NewUpdateSetRLP(updateSet, updateSet.DeletedAccounts)

	// Mock getCodeFunc
	getCodeFunc := func(codeHash types.Hash) ([]byte, error) {
		return nil, leveldb.ErrReadOnly
	}

	// Convert back to world state
	newUpdateSet, err := rlpUpdateSet.ToWorldState(getCodeFunc, 10)
	assert.Equal(t, leveldb.ErrReadOnly, err)
	assert.Nil(t, newUpdateSet)
}
