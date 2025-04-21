package updateset

import (
	pb "github.com/0xsoniclabs/substate/protobuf"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"google.golang.org/protobuf/proto"
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

func (s *UpdateSet) ToWorldStatePB() *pb.Substate_Alloc {
	return pb.ToProtobufAlloc(s.WorldState)
}

func NewUpdateSetRLP(updateSet *UpdateSet, deletedAccounts []types.Address) UpdateSetPB {
	w := updateSet.ToWorldStatePB()
	return UpdateSetPB{
		WorldState:      w,
		DeletedAccounts: deletedAccounts,
	}
}

// UpdateSetPB represents the DB structure of UpdateSet.
type UpdateSetPB struct {
	WorldState      *pb.Substate_Alloc
	DeletedAccounts []types.Address
}

func (up *UpdateSetPB) ToWorldState(getCodeFunc func(codeHash types.Hash) ([]byte, error), block uint64) (*UpdateSet, error) {
	worldState, err := up.WorldState.Decode(getCodeFunc)
	if err != nil {
		return nil, err
	}
	return NewUpdateSet(*worldState, block), nil
}

func EncodeUpdateSetPB(s *UpdateSetPB) ([]byte, error) {
	addrs := make([][]byte, 0, len(s.DeletedAccounts))
	for _, addr := range s.DeletedAccounts {
		addrs = append(addrs, addr.Bytes())
	}
	obj := &pb.UpdateSetPB{
		WorldState:      s.WorldState,
		DeletedAccounts: addrs,
	}
	return proto.Marshal(obj)
}

func DecodeUpdateSetPB(data []byte) (UpdateSetPB, error) {
	obj := &pb.UpdateSetPB{}
	err := proto.Unmarshal(data, obj)
	if err != nil {
		return UpdateSetPB{}, err
	}
	addrs := make([]types.Address, 0, len(obj.DeletedAccounts))
	for _, addr := range obj.DeletedAccounts {
		addrs = append(addrs, types.BytesToAddress(addr))
	}
	return UpdateSetPB{
		WorldState:      obj.WorldState,
		DeletedAccounts: addrs,
	}, nil
}
