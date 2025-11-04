package substate

import (
	"math/big"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"

	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/types/hash"
)

func TestMessage_EqualNonce(t *testing.T) {
	msg := &Message{Nonce: 0}
	comparedMsg := &Message{Nonce: 1}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages nonce are different but equal returned true")
	}

	comparedMsg.Nonce = msg.Nonce
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages nonce are same but equal returned false")
	}
}

func TestMessage_EqualCheckNonce(t *testing.T) {
	msg := &Message{CheckNonce: false}
	comparedMsg := &Message{CheckNonce: true}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages CheckNonce are different but equal returned true")
	}

	comparedMsg.CheckNonce = msg.CheckNonce
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages CheckNonce are same but equal returned false")
	}
}

func TestMessage_EqualGasPrice(t *testing.T) {
	msg := &Message{GasPrice: new(big.Int).SetUint64(0)}
	comparedMsg := &Message{GasPrice: new(big.Int).SetUint64(1)}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages GasPrice are different but equal returned true")
	}

	comparedMsg.GasPrice = msg.GasPrice
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages GasPrice are same but equal returned false")
	}
}

func TestMessage_EqualFrom(t *testing.T) {
	msg := &Message{From: types.Address{0}}
	comparedMsg := &Message{From: types.Address{1}}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages From are different but equal returned true")
	}

	comparedMsg.From = msg.From
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages From are same but equal returned false")
	}
}

func TestMessage_EqualTo(t *testing.T) {
	msg := &Message{To: &types.Address{0}}
	comparedMsg := &Message{To: &types.Address{1}}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages To are different but equal returned true")
	}

	comparedMsg.To = msg.To
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages To are same but equal returned false")
	}
}

func TestMessage_EqualValue(t *testing.T) {
	msg := &Message{Value: new(big.Int).SetUint64(0)}
	comparedMsg := &Message{Value: new(big.Int).SetUint64(1)}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages values are different but equal returned true")
	}

	comparedMsg.Value = msg.Value
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages Value are same but equal returned false")
	}
}

func TestMessage_Equal_DataHashDoesNotAffectResult(t *testing.T) {
	msg := &Message{dataHash: new(types.Hash)}
	*msg.dataHash = types.BytesToHash([]byte{0})
	comparedMsg := &Message{dataHash: new(types.Hash)}
	*comparedMsg.dataHash = types.BytesToHash([]byte{1})

	if !msg.Equal(comparedMsg) {
		t.Fatal("dataHash must not affect equal even if it is different")
	}

	comparedMsg.dataHash = msg.dataHash
	if !msg.Equal(comparedMsg) {
		t.Fatal("dataHash must not affect equal")
	}
}

func TestMessage_EqualAccessList(t *testing.T) {
	msg := &Message{AccessList: []types.AccessTuple{{Address: types.Address{0}, StorageKeys: []types.Hash{types.BytesToHash([]byte{0})}}}}
	comparedMsg := &Message{AccessList: []types.AccessTuple{{Address: types.Address{0}, StorageKeys: []types.Hash{types.BytesToHash([]byte{1})}}}}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages access list have different Storage Key for same address but equal returned true")
	}

	comparedMsg.AccessList = append(comparedMsg.AccessList, types.AccessTuple{Address: types.Address{0}, StorageKeys: []types.Hash{types.BytesToHash([]byte{0})}})
	if msg.Equal(comparedMsg) {
		t.Fatal("messages access list have different Storage Keys for same address but equal returned true")
	}

	comparedMsg = &Message{AccessList: []types.AccessTuple{{Address: types.Address{1}, StorageKeys: []types.Hash{types.BytesToHash([]byte{1})}}}}
	if msg.Equal(comparedMsg) {
		t.Fatal("messages access list have different AccessList but equal returned true")
	}

	comparedMsg = &Message{AccessList: []types.AccessTuple{{Address: types.Address{0},
		StorageKeys: []types.Hash{types.BytesToHash([]byte{1}), types.BytesToHash([]byte{0})},
	}}}
	if msg.Equal(comparedMsg) {
		t.Fatal("messages access list have different AccessList but equal returned true")
	}

	comparedMsg.AccessList = msg.AccessList
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages Value are same but equal returned false")
	}
}

func TestMessage_EqualGasFeeCap(t *testing.T) {
	msg := &Message{GasFeeCap: new(big.Int).SetUint64(0)}
	comparedMsg := &Message{GasFeeCap: new(big.Int).SetUint64(1)}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages GasFeeCap are different but equal returned true")
	}

	comparedMsg.GasFeeCap = msg.GasFeeCap
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages GasFeeCap are same but equal returned false")
	}
}

func TestMessage_EqualGasTipCap(t *testing.T) {
	msg := &Message{GasTipCap: new(big.Int).SetUint64(0)}
	comparedMsg := &Message{GasTipCap: new(big.Int).SetUint64(1)}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages GasTipCap are different but equal returned true")
	}

	comparedMsg.GasTipCap = msg.GasTipCap
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages GasTipCap are same but equal returned false")
	}
}

func TestMessage_EqualBlobGasFeeCap(t *testing.T) {
	msg := &Message{BlobGasFeeCap: new(big.Int).SetUint64(0)}
	comparedMsg := &Message{BlobGasFeeCap: new(big.Int).SetUint64(1)}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages BlobGasFeeCap are different but equal returned true")
	}

	comparedMsg.BlobGasFeeCap = msg.BlobGasFeeCap
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages BlobGasFeeCap are same but equal returned false")
	}
}

func TestMessage_EqualBlobHashes(t *testing.T) {
	msg := &Message{BlobHashes: []types.Hash{types.BytesToHash([]byte{0x0})}}
	comparedMsg := &Message{BlobHashes: []types.Hash{types.BytesToHash([]byte{0x1})}}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages BlobHashes are different but equal returned true")
	}

	comparedMsg.BlobHashes = msg.BlobHashes
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages BlobHashes are same but equal returned false")
	}
}

func TestMessage_DataHashReturnsIfExists(t *testing.T) {
	want := types.BytesToHash([]byte{1})
	msg := &Message{dataHash: &want}

	got, err := msg.DataHash()
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestMessage_DataHashGeneratesNewHashIfNil(t *testing.T) {
	msg := &Message{Data: []byte{1}}
	got, err := msg.DataHash()
	assert.NoError(t, err)

	want := hash.Keccak256Hash(msg.Data)

	assert.False(t, got.IsEmpty())
	assert.Equal(t, want, got)
}

func TestMessage_EqualSetCodeAuthorization(t *testing.T) {
	msg := &Message{SetCodeAuthorizations: []types.SetCodeAuthorization{{ChainID: *uint256.NewInt(1), Address: types.Address{0}, Nonce: 1, V: 1, R: *uint256.NewInt(1), S: *uint256.NewInt(1)}}}
	comparedMsg := &Message{SetCodeAuthorizations: []types.SetCodeAuthorization{{ChainID: *uint256.NewInt(2), Address: types.Address{0}, Nonce: 1, V: 1, R: *uint256.NewInt(1), S: *uint256.NewInt(1)}}}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages setCodeAuthorizations have different chainId but equal returned true")
	}
}

func TestMessage_NewMessage(t *testing.T) {
	want := &Message{
		Nonce:          1,
		CheckNonce:     true,
		GasPrice:       new(big.Int).SetUint64(1),
		Gas:            1,
		From:           types.Address{1},
		To:             &types.Address{1},
		Value:          new(big.Int).SetUint64(1),
		Data:           []byte{1},
		dataHash:       &types.Hash{},
		ProtobufTxType: new(int32),
		AccessList:     []types.AccessTuple{{Address: types.Address{1}, StorageKeys: []types.Hash{types.BytesToHash([]byte{1})}}},
		GasFeeCap:      new(big.Int).SetUint64(1),
		GasTipCap:      new(big.Int).SetUint64(1),
		BlobGasFeeCap:  new(big.Int).SetUint64(1),
		BlobHashes:     []types.Hash{types.BytesToHash([]byte{1})},
		SetCodeAuthorizations: []types.SetCodeAuthorization{
			{
				ChainID: *uint256.NewInt(1),
				Address: types.Address{1},
				Nonce:   1,
				V:       1,
				R:       *uint256.NewInt(1),
				S:       *uint256.NewInt(1),
			},
		},
	}

	got := NewMessage(
		1,
		true,
		new(big.Int).SetUint64(1),
		1,
		types.Address{1},
		&types.Address{1},
		new(big.Int).SetUint64(1),
		[]byte{1},
		&types.Hash{},
		new(int32),
		[]types.AccessTuple{{Address: types.Address{1}, StorageKeys: []types.Hash{types.BytesToHash([]byte{1})}}},
		new(big.Int).SetUint64(1),
		new(big.Int).SetUint64(1),
		new(big.Int).SetUint64(1),
		[]types.Hash{types.BytesToHash([]byte{1})},
		[]types.SetCodeAuthorization{{ChainID: *uint256.NewInt(1), Address: types.Address{1}, Nonce: 1, V: 1, R: *uint256.NewInt(1), S: *uint256.NewInt(1)}},
	)

	assert.Equal(t, want, got)
}

func TestMessage_EqualSelf(t *testing.T) {
	msg := &Message{Nonce: 1}
	assert.Equal(t, true, msg.Equal(msg))
}

func TestMessage_EqualNil(t *testing.T) {
	msg := &Message{Nonce: 1}
	assert.Equal(t, false, msg.Equal(nil))
}
func TestMessage_String(t *testing.T) {
	msg := &Message{
		Nonce:          1,
		CheckNonce:     true,
		GasPrice:       new(big.Int).SetUint64(1),
		Gas:            1,
		From:           types.Address{1},
		To:             &types.Address{1},
		Value:          new(big.Int).SetUint64(1),
		Data:           []byte{1},
		dataHash:       &types.Hash{},
		ProtobufTxType: new(int32),
		AccessList:     []types.AccessTuple{{Address: types.Address{1}, StorageKeys: []types.Hash{types.BytesToHash([]byte{1})}}},
		GasFeeCap:      new(big.Int).SetUint64(1),
		GasTipCap:      new(big.Int).SetUint64(1),
		BlobGasFeeCap:  new(big.Int).SetUint64(1),
		BlobHashes:     []types.Hash{types.BytesToHash([]byte{1})},
		SetCodeAuthorizations: []types.SetCodeAuthorization{
			{
				ChainID: *uint256.NewInt(1),
				Address: types.Address{1},
				Nonce:   1,
				V:       1,
				R:       *uint256.NewInt(1),
				S:       *uint256.NewInt(1),
			},
		},
	}

	expected := "Nonce: 1\nCheckNonce: true\nFrom: 0x0100000000000000000000000000000000000000\nTo: 0x0100000000000000000000000000000000000000\nValue: 1\nData: \x01\nData Hash: 0x0000000000000000000000000000000000000000000000000000000000000000\nGas Fee Cap: 1\nGas Tip Cap: 1\nAddress: 0x0100000000000000000000000000000000000000Storage Key 0: 0x0000000000000000000000000000000000000000000000000000000000000001SetCodeAuthorization:\nChainID: 1\nAddress: 0x0100000000000000000000000000000000000000\nNonce: 1\nV: 1\nR: 1\nS: 1\n"
	assert.Equal(t, expected, msg.String())
}
