package rlp

import (
	"math/big"
	"testing"

	"github.com/OxSonicLabs/Substate/substate"
	"github.com/OxSonicLabs/Substate/types/hash"
)

func TestNewMessage_InitCodeHashIsCreated_WhenToIsNil(t *testing.T) {
	data := []byte{0x1}
	m := NewMessage(&substate.Message{Data: data, Value: big.NewInt(1), To: nil})
	if got, want := *m.InitCodeHash, hash.Keccak256Hash(data); got != want {
		t.Fatalf("unexpected code hash\ngot: %s\nwant: %s", got, want)
	}
}
