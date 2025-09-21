package updateset

import (
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
)

func NewUpdateSet(alloc substate.WorldState, block uint64) *UpdateSet {
	return &UpdateSet{
		WorldState: alloc,
		Block:      block,
	}
}

// UpdateSet represents the substate.Account world state for the block.
type UpdateSet struct {
	WorldState      substate.WorldState
	Block           uint64
	DeletedAccounts []types.Address
}

func (s *UpdateSet) Equal(y *UpdateSet) bool {
	if s == y {
		return true
	}
	if !s.WorldState.Equal(y.WorldState) {
		return false
	}

	if s.Block != y.Block {
		return false
	}

	for i, val := range s.DeletedAccounts {
		if val != y.DeletedAccounts[i] {
			return false
		}
	}
	return true
}
