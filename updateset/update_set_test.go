package updateset

import (
	"testing"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

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
