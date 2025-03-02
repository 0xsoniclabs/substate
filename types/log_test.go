package types

import (
	"testing"

	"github.com/0xsoniclabs/substate/types/rlp"
	"github.com/stretchr/testify/assert"
)

func TestLog_EncodeRLPWithEmptyLog(t *testing.T) {
	log := &Log{}
	bytes, err := rlp.EncodeToBytes(log)
	assert.Nil(t, err)

	var decoded Log
	err = rlp.DecodeBytes(bytes, &decoded)
	assert.Nil(t, err)
	assert.Equal(t, log.Address, decoded.Address)
	assert.Equal(t, []Hash{}, decoded.Topics)
	assert.Equal(t, []byte{}, decoded.Data)
}

func TestLog_EncodeRLPWithFullLog(t *testing.T) {
	log := &Log{
		Address: Address{1, 2, 3},
		Topics:  []Hash{{4, 5, 6}, {7, 8, 9}},
		Data:    []byte{10, 11, 12},
	}

	bytes, err := rlp.EncodeToBytes(log)
	assert.Nil(t, err)

	var decoded Log
	err = rlp.DecodeBytes(bytes, &decoded)
	assert.Nil(t, err)
	assert.Equal(t, log.Address, decoded.Address)
	assert.Equal(t, log.Topics, decoded.Topics)
	assert.Equal(t, log.Data, decoded.Data)
}

func TestLog_DecodeRLPWithInvalidData(t *testing.T) {
	var log Log
	err := rlp.DecodeBytes([]byte{0x01}, &log)
	assert.NotNil(t, err)
}

func TestLog_EncodeRLPPreservesConsensusFields(t *testing.T) {
	original := &Log{
		Address:     Address{1},
		Topics:      []Hash{{2}},
		Data:        []byte{3},
		BlockNumber: 4,
		TxHash:      Hash{5},
		TxIndex:     6,
		BlockHash:   Hash{7},
		Index:       8,
		Removed:     true,
	}

	bytes, err := rlp.EncodeToBytes(original)
	assert.Nil(t, err)

	var decoded Log
	err = rlp.DecodeBytes(bytes, &decoded)
	assert.Nil(t, err)

	// Only consensus fields should be encoded/decoded
	assert.Equal(t, original.Address, decoded.Address)
	assert.Equal(t, original.Topics, decoded.Topics)
	assert.Equal(t, original.Data, decoded.Data)

	// Non-consensus fields should be zero values
	assert.Equal(t, uint64(0), decoded.BlockNumber)
	assert.Equal(t, Hash{}, decoded.TxHash)
	assert.Equal(t, uint(0), decoded.TxIndex)
	assert.Equal(t, Hash{}, decoded.BlockHash)
	assert.Equal(t, uint(0), decoded.Index)
	assert.Equal(t, false, decoded.Removed)
}
