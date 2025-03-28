package db

import (
	"fmt"

	"github.com/0xsoniclabs/substate/protobuf"
	"github.com/0xsoniclabs/substate/types"
	trlp "github.com/0xsoniclabs/substate/types/rlp"
	"google.golang.org/protobuf/proto"
)

func (db *destroyedAccountDB) GetSubstateEncoding() SubstateEncodingSchema {
	return db.encoding.schema
}

func (db *destroyedAccountDB) SetSubstateEncoding(schema SubstateEncodingSchema) error {
	encoding, err := newDestroyedAccountEncoding(schema)
	if err != nil {
		return fmt.Errorf("failed to set decodeFunc; %w", err)
	}

	db.encoding = *encoding
	return nil
}

type destroyedAccountEncoding struct {
	schema SubstateEncodingSchema
	encode func(list SuicidedAccountLists) ([]byte, error)
	decode func(data []byte) (SuicidedAccountLists, error)
}

func newDestroyedAccountEncoding(encoding SubstateEncodingSchema) (*destroyedAccountEncoding, error) {
	switch encoding {
	case DefaultEncodingSchema, ProtobufEncodingSchema:
		return &destroyedAccountEncoding{
			schema: ProtobufEncodingSchema,
			encode: encodeSuicidedAccountListPB,
			decode: decodeSuicidedAccountListPB,
		}, nil
	case RLPEncodingSchema:
		return &destroyedAccountEncoding{
			schema: RLPEncodingSchema,
			encode: encodeSuicidedAccountListRLP,
			decode: decodeSuicidedAccountListRLP,
		}, nil
	default:
		return nil, fmt.Errorf("encoding not supported: %s", encoding)
	}
}

func encodeSuicidedAccountListPB(list SuicidedAccountLists) ([]byte, error) {
	pbAccountList := &protobuf.SuicidedAccountLists{}
	for _, addr := range list.DestroyedAccounts {
		pbAccountList.DestroyedAccounts = append(pbAccountList.DestroyedAccounts, addr.Bytes())
	}
	for _, addr := range list.ResurrectedAccounts {
		pbAccountList.ResurrectedAccounts = append(pbAccountList.ResurrectedAccounts, addr.Bytes())
	}
	return proto.Marshal(pbAccountList)
}

func decodeSuicidedAccountListPB(data []byte) (SuicidedAccountLists, error) {
	pbAccountList := &protobuf.SuicidedAccountLists{}
	err := proto.Unmarshal(data, pbAccountList)
	if err != nil {
		return SuicidedAccountLists{}, err
	}
	list := SuicidedAccountLists{}
	for _, addr := range pbAccountList.DestroyedAccounts {
		list.DestroyedAccounts = append(list.DestroyedAccounts, types.BytesToAddress(addr))
	}
	for _, addr := range pbAccountList.ResurrectedAccounts {
		list.ResurrectedAccounts = append(list.ResurrectedAccounts, types.BytesToAddress(addr))
	}
	return list, nil
}

func encodeSuicidedAccountListRLP(list SuicidedAccountLists) ([]byte, error) {
	return trlp.EncodeToBytes(list)
}

func decodeSuicidedAccountListRLP(data []byte) (SuicidedAccountLists, error) {
	list := SuicidedAccountLists{}
	err := trlp.DecodeBytes(data, &list)
	return list, err
}
