package rlp

import (
	"math/big"
	"testing"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types/hash"
	"github.com/stretchr/testify/assert"
)

func TestNewMessage_InitCodeHashIsCreated_WhenToIsNil(t *testing.T) {
	data := []byte{0x1}
	m, err := NewMessage(&substate.Message{Data: data, Value: big.NewInt(1), To: nil})
	assert.NoError(t, err)
	assert.Equal(t, hash.Keccak256Hash(data), *m.InitCodeHash)
}
