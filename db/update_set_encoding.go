package db

import (
	"fmt"

	"github.com/0xsoniclabs/substate/protobuf"
	"github.com/0xsoniclabs/substate/rlp"
	"github.com/0xsoniclabs/substate/types"
	trlp "github.com/0xsoniclabs/substate/types/rlp"
	"github.com/0xsoniclabs/substate/updateset"
	"google.golang.org/protobuf/proto"
)

func (db *updateDB) GetSubstateEncoding() SubstateEncodingSchema {
	return db.encoding.schema
}

func (db *updateDB) SetSubstateEncoding(schema SubstateEncodingSchema) error {
	encoding, err := newUpdateSetEncoding(schema)
	if err != nil {
		return fmt.Errorf("failed to set decoder; %w", err)
	}

	db.encoding = *encoding
	return nil
}

type UpdateSetEncoderFunc = func(updateSet updateset.UpdateSet, deletedAccounts []types.Address) ([]byte, error)
type UpdateSetDecoderFunc = func(block uint64, getCode func(codeHash types.Hash) ([]byte, error), data []byte) (*updateset.UpdateSet, error)
type updateSetEncoding struct {
	schema SubstateEncodingSchema
	encode UpdateSetEncoderFunc
	decode UpdateSetDecoderFunc
}

func newUpdateSetEncoding(encoding SubstateEncodingSchema) (*updateSetEncoding, error) {
	switch encoding {
	case DefaultEncodingSchema, ProtobufEncodingSchema:
		return &updateSetEncoding{
			schema: ProtobufEncodingSchema,
			encode: encodeUpdateSetPB,
			decode: decodeUpdateSetPB,
		}, nil
	case RLPEncodingSchema:
		return &updateSetEncoding{
			schema: RLPEncodingSchema,
			encode: encodeUpdateSetRLP,
			decode: decodeUpdateSetRLP,
		}, nil
	default:
		return nil, fmt.Errorf("encoding not supported: %s", encoding)
	}
}

func encodeUpdateSetPB(updateSet updateset.UpdateSet, deletedAccounts []types.Address) ([]byte, error) {
	up, err := protobuf.NewUpdateSetPB(updateSet.WorldState, deletedAccounts)
	if err != nil {
		return nil, err
	}
	addrs := make([][]byte, 0, len(up.DeletedAccounts))
	for _, addr := range up.DeletedAccounts {
		addrs = append(addrs, addr.Bytes())
	}
	obj := &protobuf.UpdateSet{
		WorldState:      up.WorldState,
		DeletedAccounts: addrs,
	}
	return proto.Marshal(obj)
}

func decodeUpdateSetPB(block uint64, getCode func(codeHash types.Hash) ([]byte, error), data []byte) (*updateset.UpdateSet, error) {
	obj := &protobuf.UpdateSet{}
	err := proto.Unmarshal(data, obj)
	if err != nil {
		return nil, err
	}
	addrs := make([]types.Address, 0, len(obj.DeletedAccounts))
	for _, addr := range obj.DeletedAccounts {
		addrs = append(addrs, types.BytesToAddress(addr))
	}
	up := protobuf.UpdateSetPB{
		WorldState:      obj.WorldState,
		DeletedAccounts: addrs,
	}
	ws, err := up.ToWorldState(getCode)
	if err != nil {
		return nil, err
	}
	return updateset.NewUpdateSet(*ws, block), nil
}

func encodeUpdateSetRLP(updateSet updateset.UpdateSet, deletedAccounts []types.Address) ([]byte, error) {
	up, err := rlp.NewUpdateSetRLP(updateSet.WorldState, deletedAccounts)
	if err != nil {
		return nil, err
	}
	value, err := trlp.EncodeToBytes(up)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func decodeUpdateSetRLP(block uint64, getCode func(codeHash types.Hash) ([]byte, error), data []byte) (*updateset.UpdateSet, error) {
	var up rlp.UpdateSetRLP
	if err := trlp.DecodeBytes(data, &up); err != nil {
		return nil, err
	}
	ws, err := up.ToWorldState(getCode)
	if err != nil {
		return nil, err
	}
	return updateset.NewUpdateSet(*ws, block), nil
}
