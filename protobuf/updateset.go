package protobuf

import (
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
)

type UpdateSetPB struct {
	WorldState      *Alloc
	DeletedAccounts []types.Address
}

func NewUpdateSetPB(ws substate.WorldState, deletedAccounts []types.Address) *UpdateSetPB {
	return &UpdateSetPB{
		WorldState:      toProtobufAlloc(ws),
		DeletedAccounts: deletedAccounts,
	}
}
