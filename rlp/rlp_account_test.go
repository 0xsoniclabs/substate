package rlp

import (
	"testing"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestNewRLPAccount(t *testing.T) {
	// given
	acc := &substate.Account{
		Nonce:   5,
		Balance: uint256.NewInt(1000),
		Storage: map[types.Hash]types.Hash{
			{0x03}: {0x04},
			{0x01}: {0x02},
		},
		Code: []byte{0x60, 0x60, 0x60},
	}

	// when
	result := NewRLPAccount(acc)

	// then
	assert.NotNil(t, result)
	assert.Equal(t, uint64(5), result.Nonce)
	assert.Equal(t, uint256.NewInt(1000), result.Balance)
	assert.NotEmpty(t, result.CodeHash)
	assert.Len(t, result.Storage, 2)
	// verify storage is sorted
	assert.True(t, result.Storage[0][0].Compare(result.Storage[1][0]) < 0)
}
