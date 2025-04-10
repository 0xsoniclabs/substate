package updateset

import (
	"encoding/binary"
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

	// Convert to PB
	pbState := updateSet.ToWorldStatePB()

	// Verify word state length
	assert.Equal(t, len(ws), len(pbState.Alloc))

	// Find the accounts in the RLP representation
	var foundAcc1, foundAcc2 bool
	for _, alloc := range pbState.Alloc {
		acc := alloc.Account
		addr := types.BytesToAddress(alloc.Address)

		if addr == (types.Address{1}) {
			foundAcc1 = true
			assert.Equal(t, uint64(1), acc.GetNonce())
			assert.Equal(t, uint64(100), bytesToUint64(acc.GetBalance()))
			assert.Equal(t, acc1.CodeHash(), types.BytesToHash(acc.GetCodeHash()))
			assert.Equal(t, 1, len(acc.GetStorage()))
		}

		if addr == (types.Address{2}) {
			foundAcc2 = true
			assert.Equal(t, uint64(2), acc.GetNonce())
			assert.Equal(t, uint64(200), bytesToUint64(acc.GetBalance()))
			assert.Equal(t, acc2.CodeHash(), types.BytesToHash(acc.GetCodeHash()))
			assert.Equal(t, 0, len(acc.GetStorage()))
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
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "leveldb: read-only mode")
	assert.Nil(t, newUpdateSet)
}

// bytesToUint64 converts a byte slice to a uint64 value.
func bytesToUint64(b []byte) uint64 {
	// Ensure the byte slice has enough bytes
	if len(b) < 8 {
		// Create a new 8-byte slice
		buf := make([]byte, 8)
		// Copy the input bytes to the end of the buffer
		copy(buf[8-len(b):], b)
		b = buf
	}
	return binary.BigEndian.Uint64(b)
}
