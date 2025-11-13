// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

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
