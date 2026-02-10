package db

import (
	"encoding/hex"
	"testing"

	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDestroyedAccountDB_GetSubstateEncoding(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	actual := db.GetSubstateEncoding()
	assert.Equal(t, RLPEncodingSchema, actual)
}

func TestDestroyedAccountDB_SetSubstateEncoding(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)
	// Test setting to RLP
	err := db.SetSubstateEncoding(RLPEncodingSchema)
	assert.Nil(t, err)
	assert.Equal(t, RLPEncodingSchema, db.GetSubstateEncoding())

	// Test setting to Protobuf (should error)
	err = db.SetSubstateEncoding(ProtobufEncodingSchema)
	assert.Error(t, err)
	assert.Equal(t, "failed to set decoder; protobuf encoding is disabled for destroyed account db", err.Error())
	assert.Equal(t, RLPEncodingSchema, db.GetSubstateEncoding()) // Should remain RLP

	// Test setting to Default (should map to RLP)
	err = db.SetSubstateEncoding(DefaultEncodingSchema)
	assert.Nil(t, err)
	assert.Equal(t, RLPEncodingSchema, db.GetSubstateEncoding())

	err = db.SetSubstateEncoding("invalid")
	assert.Error(t, err)
}

func TestDestroyedAccountDB_Encode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	actual, err := db.Encode(SuicidedAccountLists{})
	assert.NoError(t, err)
	// RLP encoding header
	// 0xc2 = list prefix for 2 bytes
	// 0xc0 = empty list (DestroyedAccounts)
	// 0xc0 = empty list (ResurrectedAccounts)
	assert.Equal(t, []uint8{0xc2, 0xc0, 0xc0}, actual)
}

func TestDestroyedAccountDB_Decode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := NewMockDbAdapter(ctrl)
	db := newTestDestroyedAccountDB(t, baseDb, DefaultEncodingSchema)

	actual, err := db.Decode([]uint8{0xc2, 0xc0, 0xc0})
	assert.NoError(t, err)
	assert.Equal(t, SuicidedAccountLists{
		DestroyedAccounts:   []types.Address{},
		ResurrectedAccounts: []types.Address{},
	}, actual)
}

func TestDestroyedAccountEncoding_newDestroyedAccountEncoding(t *testing.T) {
	encoding, err := newDestroyedAccountEncoding(DefaultEncodingSchema)
	assert.NoError(t, err)
	assert.Equal(t, RLPEncodingSchema, encoding.schema)

	encoding, err = newDestroyedAccountEncoding(ProtobufEncodingSchema)
	assert.Error(t, err)
	assert.Nil(t, encoding)

	encoding, err = newDestroyedAccountEncoding(RLPEncodingSchema)
	assert.NoError(t, err)
	assert.Equal(t, RLPEncodingSchema, encoding.schema)

	encoding, err = newDestroyedAccountEncoding("unsupported")
	assert.Error(t, err)
	assert.Nil(t, encoding)
}

func TestDestroyedAccountEncoding_encodeSuicidedAccountListPB(t *testing.T) {
	input := SuicidedAccountLists{
		DestroyedAccounts:   []types.Address{{1}, {2}},
		ResurrectedAccounts: []types.Address{},
	}

	output, err := encodeSuicidedAccountListPB(input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "644b98e4aaa1be108141724dee726191", bytesToMD5(output))
}

func TestDestroyedAccountEncoding_decodeSuicidedAccountListPB(t *testing.T) {
	expected := SuicidedAccountLists{
		DestroyedAccounts:   []types.Address{{1}, {2}},
		ResurrectedAccounts: []types.Address(nil),
	}

	input, err := hex.DecodeString("0a1401000000000000000000000000000000000000000a140200000000000000000000000000000000000000")
	require.NoError(t, err)

	output, err := decodeSuicidedAccountListPB(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, output)
}

func TestDestroyedAccountEncoding_encodeSuicidedAccountListRLP(t *testing.T) {
	input := SuicidedAccountLists{
		DestroyedAccounts:   []types.Address{{1}, {2}},
		ResurrectedAccounts: []types.Address{},
	}

	output, err := encodeSuicidedAccountListRLP(input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "0b67f4149fafe9d3bbfeab050c372e58", bytesToMD5(output))
}

func TestDestroyedAccountEncoding_decodeSuicidedAccountListRLP(t *testing.T) {
	expected := SuicidedAccountLists{
		DestroyedAccounts:   []types.Address{{1}, {2}},
		ResurrectedAccounts: []types.Address{},
	}

	input, err := hex.DecodeString("ecea940100000000000000000000000000000000000000940200000000000000000000000000000000000000c0")
	require.NoError(t, err)

	output, err := decodeSuicidedAccountListRLP(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, output)
}
