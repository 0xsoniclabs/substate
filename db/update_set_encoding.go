package db

import (
	"errors"
	"fmt"

	"github.com/0xsoniclabs/substate/protobuf"
	"github.com/0xsoniclabs/substate/rlp"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	trlp "github.com/0xsoniclabs/substate/types/rlp"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/syndtr/goleveldb/leveldb"
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
	s := protobuf.NewUpdateSetPB(updateSet.WorldState, deletedAccounts)
	addrs := make([][]byte, 0, len(s.DeletedAccounts))
	for _, addr := range s.DeletedAccounts {
		addrs = append(addrs, addr.Bytes())
	}
	obj := &protobuf.UpdateSet{
		WorldState:      s.WorldState,
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
	worldState, err := up.WorldState.Decode(getCode)
	if err != nil {
		return nil, err
	}
	return updateset.NewUpdateSet(*worldState, block), nil
}

func encodeUpdateSetRLP(updateSet updateset.UpdateSet, deletedAccounts []types.Address) ([]byte, error) {
	ws := rlp.WorldState{
		Addresses: []types.Address{},
		Accounts:  []*rlp.SubstateAccountRLP{},
	}

	for addr, acc := range updateSet.WorldState {
		ws.Addresses = append(ws.Addresses, addr)
		ws.Accounts = append(ws.Accounts, rlp.NewRLPAccount(acc))
	}

	up := rlp.UpdateSetRLP{
		WorldState:      ws,
		DeletedAccounts: deletedAccounts,
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
	worldState := make(substate.WorldState)
	for i, addr := range up.WorldState.Addresses {
		worldStateAcc := up.WorldState.Accounts[i]

		code, err := getCode(worldStateAcc.CodeHash)
		if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
			return nil, err
		}

		acc := substate.Account{
			Nonce:   worldStateAcc.Nonce,
			Balance: worldStateAcc.Balance,
			Storage: make(map[types.Hash]types.Hash),
			Code:    code,
		}

		for j := range worldStateAcc.Storage {
			acc.Storage[worldStateAcc.Storage[j][0]] = worldStateAcc.Storage[j][1]
		}
		worldState[addr] = &acc
	}

	return updateset.NewUpdateSet(worldState, block), nil
}
