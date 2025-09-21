package db

import (
	"crypto/md5"
	"encoding/hex"
	"testing"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUpdateDB_GetSubstateEncoding(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)
	db := newTestUpdateDB(t, mockDB, ProtobufEncodingSchema)

	encoding := db.GetSubstateEncoding()
	assert.Equal(t, ProtobufEncodingSchema, encoding)
}

func TestUpdateDB_SetSubstateEncoding(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockCodeDB(ctrl)
	db := newTestUpdateDB(t, mockDB, ProtobufEncodingSchema)

	// Test setting to RLP
	err := db.SetSubstateEncoding(RLPEncodingSchema)
	assert.Nil(t, err)
	assert.Equal(t, RLPEncodingSchema, db.encoding.schema)

	// Test setting to Protobuf
	err = db.SetSubstateEncoding(ProtobufEncodingSchema)
	assert.Nil(t, err)
	assert.Equal(t, ProtobufEncodingSchema, db.encoding.schema)

	// Test setting to Default (should map to Protobuf)
	err = db.SetSubstateEncoding(DefaultEncodingSchema)
	assert.Nil(t, err)
	assert.Equal(t, ProtobufEncodingSchema, db.encoding.schema)

	// Test setting to an unknown schema
	err = db.SetSubstateEncoding(SubstateEncodingSchema("unknown"))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "encoding not supported")
}

func TestUpdateSetEncoding_newUpdateSetEncoding(t *testing.T) {
	encoding, err := newUpdateSetEncoding(DefaultEncodingSchema)
	assert.NoError(t, err)
	assert.Equal(t, ProtobufEncodingSchema, encoding.schema)

	encoding, err = newUpdateSetEncoding(ProtobufEncodingSchema)
	assert.NoError(t, err)
	assert.Equal(t, ProtobufEncodingSchema, encoding.schema)

	encoding, err = newUpdateSetEncoding(RLPEncodingSchema)
	assert.NoError(t, err)
	assert.Equal(t, RLPEncodingSchema, encoding.schema)

	encoding, err = newUpdateSetEncoding("unsupported")
	assert.Error(t, err)
	assert.Nil(t, encoding)
}

func TestUpdateSetEncoding_encodeUpdateSetPB(t *testing.T) {
	updateSet := updateset.UpdateSet{
		WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
		Block:           0,
		DeletedAccounts: []types.Address{},
	}
	deletedAccounts := []types.Address{{}}
	value, err := encodeUpdateSetPB(updateSet, deletedAccounts)

	assert.NoError(t, err)
	assert.Equal(t, "06753c366fe2f1b1bdc6d67cb3e3698f", bytesToMD5(value))
}

func TestUpdateSetEncoding_decodeUpdateSetPB(t *testing.T) {
	expected := &updateset.UpdateSet{
		WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
		Block:           0,
		DeletedAccounts: []types.Address(nil),
	}

	input := []byte{0xa, 0x41, 0xa, 0x3f, 0xa, 0x14, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x12, 0x27, 0x8, 0x1, 0x12, 0x1, 0x1, 0x2a, 0x20, 0xc5, 0xd2, 0x46, 0x1, 0x86, 0xf7, 0x23, 0x3c, 0x92, 0x7e, 0x7d, 0xb2, 0xdc, 0xc7, 0x3, 0xc0, 0xe5, 0x0, 0xb6, 0x53, 0xca, 0x82, 0x27, 0x3b, 0x7b, 0xfa, 0xd8, 0x4, 0x5d, 0x85, 0xa4, 0x70, 0x12, 0x14, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

	value, err := decodeUpdateSetPB(0, func(codeHash types.Hash) ([]byte, error) {
		return nil, nil
	}, input)
	assert.NoError(t, err)
	assert.Equal(t, expected.WorldState, value.WorldState)
	assert.Equal(t, expected.Block, value.Block)
	assert.Equal(t, expected.DeletedAccounts, value.DeletedAccounts)
}

func TestUpdateSetEncoding_encodeUpdateSetRLP(t *testing.T) {
	updateSet := updateset.UpdateSet{
		WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
		Block:           0,
		DeletedAccounts: []types.Address{},
	}
	deletedAccounts := []types.Address{{}}
	value, err := encodeUpdateSetRLP(updateSet, deletedAccounts)

	assert.NoError(t, err)
	// byte to hex string
	t.Log(hex.EncodeToString(value))
	assert.Equal(t, "a6aa2512c0347704d3a9a302fc0b0fa1", bytesToMD5(value))

}

func TestUpdateSetEncoding_decodeUpdateSetRLP(t *testing.T) {
	expected := &updateset.UpdateSet{
		WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
		Block:           0,
		DeletedAccounts: []types.Address(nil),
	}

	input, err := hex.DecodeString("f854f83cd5940100000000000000000000000000000000000000e5e40101a0c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470c0d5940000000000000000000000000000000000000000")
	require.NoError(t, err)
	value, err := decodeUpdateSetRLP(0, func(codeHash types.Hash) ([]byte, error) {
		return nil, nil
	}, input)
	assert.NoError(t, err)
	assert.Equal(t, expected.WorldState, value.WorldState)
	assert.Equal(t, expected.Block, value.Block)
	assert.Equal(t, expected.DeletedAccounts, value.DeletedAccounts)
}

// bytesToMD5 computes the MD5 hash of the input byte slice and returns it as a hexadecimal string.
func bytesToMD5(data []byte) string {
	hash := md5.Sum(data)
	hashString := hex.EncodeToString(hash[:])
	return hashString
}
