package protobuf

import (
	"errors"
	"testing"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestNewUpdateSetPB(t *testing.T) {
	// given
	worldState := substate.WorldState{
		types.Address{0x01}: &substate.Account{
			Nonce:   1,
			Balance: uint256.NewInt(1000),
			Storage: map[types.Hash]types.Hash{
				{0x01}: {0x02},
			},
			Code: []byte{0x03},
		},
	}
	deletedAccounts := []types.Address{{0x02}}

	// when
	result, err := NewUpdateSetPB(worldState, deletedAccounts)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.WorldState)
	assert.Equal(t, deletedAccounts, result.DeletedAccounts)
	assert.Len(t, result.WorldState.Alloc, len(worldState))
}

func TestUpdateSetPB_ToWorldStateSuccess(t *testing.T) {
	// given
	nonce := uint64(1)
	updateSet := &UpdateSetPB{
		WorldState: &Alloc{
			Alloc: []*AllocEntry{
				{
					Address: types.Address{0x01}.Bytes(),
					Account: &Account{
						Nonce:   &nonce,
						Balance: uint256.NewInt(1000).Bytes(),
						Storage: []*Account_StorageEntry{},
						Contract: &Account_CodeHash{
							CodeHash: types.Hash{0x01}.Bytes(),
						},
					},
				},
			},
		},
		DeletedAccounts: []types.Address{},
	}
	lookup := func(codeHash types.Hash) ([]byte, error) {
		return []byte{0x60}, nil
	}

	// when
	result, err := updateSet.ToWorldState(lookup)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, *result, len(updateSet.WorldState.Alloc))
}

func TestUpdateSetPB_ToWorldStateLookupError(t *testing.T) {
	// given
	nonce := uint64(1)
	updateSet := &UpdateSetPB{
		WorldState: &Alloc{
			Alloc: []*AllocEntry{
				{
					Address: types.Address{0x01}.Bytes(),
					Account: &Account{
						Nonce:   &nonce,
						Balance: uint256.NewInt(1000).Bytes(),
						Storage: []*Account_StorageEntry{},
						Contract: &Account_CodeHash{
							CodeHash: types.Hash{0x01}.Bytes(),
						},
					},
				},
			},
		},
		DeletedAccounts: []types.Address{},
	}
	lookup := func(codeHash types.Hash) ([]byte, error) {
		return nil, errors.New("lookup failed")
	}

	// when
	result, err := updateSet.ToWorldState(lookup)

	// then
	assert.Error(t, err)
	assert.Nil(t, result)
}
