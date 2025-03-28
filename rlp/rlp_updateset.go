package rlp

import (
	"github.com/0xsoniclabs/substate/types"
)

type UpdateSetRLP struct {
	WorldState      WorldState
	DeletedAccounts []types.Address
}
