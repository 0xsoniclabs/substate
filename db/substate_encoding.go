package db

import (
	"fmt"

	pb "github.com/0xsoniclabs/substate/protobuf"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/golang/protobuf/proto"
)

// SetSubstateEncoding sets the runtime encoding/decoding behavior of substateDB
// intended usage:
//
//	db := &substateDB{..} // default to rlp
//	     err := db.SetSubstateEncoding(<schema>) // set encoding
//	     db.GetSubstateDecoder() // returns configured encoding
func (db *substateDB) SetSubstateEncoding(schema string) error {
	encoding, err := newSubstateEncoding(schema, db.GetCode)
	if err != nil {
		return fmt.Errorf("failed to set decoder; %w", err)
	}

	db.encoding = encoding
	return nil
}

// GetDecoder returns the encoding in use
func (db *substateDB) GetSubstateEncoding() string {
	if db.encoding == nil {
		return ""
	}
	return db.encoding.schema
}

type substateEncoding struct {
	schema string
	decode decodeFunc
	encode encodeFunc
}

// decodeFunc aliases the common function used to decode substate
type decodeFunc func([]byte, uint64, int) (*substate.Substate, error)

// encodeFunc alias the common function used to encode substate
type encodeFunc func(*substate.Substate, uint64, int) ([]byte, error)

// codeLookupFunc aliases codehash->code lookup necessary to decode substate
type codeLookupFunc = func(types.Hash) ([]byte, error)

// newSubstateDecoder returns requested SubstateDecoder
func newSubstateEncoding(encoding string, lookup codeLookupFunc) (*substateEncoding, error) {
	switch encoding {
	case "", "default", "protobuf", "pb":
		return &substateEncoding{
			schema: "protobuf",
			decode: func(bytes []byte, block uint64, tx int) (*substate.Substate, error) {
				return decodeProtobuf(bytes, lookup, block, tx)
			},
			encode: pb.Encode,
		}, nil

	default:
		return nil, fmt.Errorf("encoding not supported: %s", encoding)

	}
}

// decodeSubstate defensively defaults to "default" if nil
func (db *substateDB) decodeToSubstate(bytes []byte, block uint64, tx int) (*substate.Substate, error) {
	return db.encoding.decode(bytes, block, tx)
}

// encodeSubstate defensively defaults to "default" if nil
func (db *substateDB) encodeSubstate(ss *substate.Substate, block uint64, tx int) ([]byte, error) {
	return db.encoding.encode(ss, block, tx)
}

// decodeProtobuf decodes protobuf-encoded bytecode into substate
func decodeProtobuf(bytes []byte, lookup codeLookupFunc, block uint64, tx int) (*substate.Substate, error) {
	pbSubstate := &pb.Substate{}
	if err := proto.Unmarshal(bytes, pbSubstate); err != nil {
		return nil, fmt.Errorf("cannot decode substate data from protobuf block: %v, tx %v; %w", block, tx, err)
	}

	return pbSubstate.Decode(lookup, block, tx)
}
