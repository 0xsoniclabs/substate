package protobuf

import (
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
)

type UpdateSetPB struct {
	WorldState      *Alloc
	DeletedAccounts []types.Address
}

func NewUpdateSetPB(ws substate.WorldState, deletedAccounts []types.Address) (*UpdateSetPB, error) {
	data, err := toProtobufAlloc(ws)
	if err != nil {
		return nil, err
	}
	return &UpdateSetPB{
		WorldState:      data,
		DeletedAccounts: deletedAccounts,
	}, nil
}

func (up *UpdateSetPB) ToWorldState(lookup func(codeHash types.Hash) ([]byte, error)) (*substate.WorldState, error) {
	return up.WorldState.decode(lookup)
}
