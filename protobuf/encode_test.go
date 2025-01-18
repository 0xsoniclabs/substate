package protobuf

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"math/big"
	"testing"
)

func bytesToMD5(data []byte) string {
	hash := md5.Sum(data)
	hashString := hex.EncodeToString(hash[:])
	return hashString
}

func TestEncodeSubstate(t *testing.T) {
	// given
	input := &substate.Substate{
		InputSubstate: substate.WorldState{
			types.Address{0x01}: &substate.Account{
				Nonce:   1,
				Balance: big.NewInt(1000),
				Storage: map[types.Hash]types.Hash{
					{0x01}: {0x02},
				},
				Code: []byte{0x03},
			},
		},
		OutputSubstate: substate.WorldState{
			types.Address{0x04}: &substate.Account{
				Nonce:   1,
				Balance: big.NewInt(2000),
				Storage: map[types.Hash]types.Hash{
					{0xCD}: {0xAB},
				},
				Code: []byte{0x07},
			},
		},
		Env: &substate.Env{
			Coinbase:    types.Address{0x01},
			GasLimit:    1000000,
			Number:      1,
			Timestamp:   1633024800,
			BlockHashes: map[uint64]types.Hash{1: {0x02}},
			BaseFee:     big.NewInt(1000),
			BlobBaseFee: big.NewInt(2000),
			Difficulty:  big.NewInt(3000),
			Random:      &types.Hash{0x03},
		},
		Message: &substate.Message{
			Nonce:          1,
			CheckNonce:     true,
			GasPrice:       big.NewInt(100),
			Gas:            21000,
			From:           types.Address{0x04},
			To:             &types.Address{0x05},
			Value:          big.NewInt(500),
			Data:           []byte{0x06},
			ProtobufTxType: proto.Int32(0),
			AccessList: []types.AccessTuple{
				{
					Address: types.Address{0x07},
					StorageKeys: []types.Hash{
						{0x08},
					},
				},
			},
			GasFeeCap:     big.NewInt(1000),
			GasTipCap:     big.NewInt(2000),
			BlobGasFeeCap: big.NewInt(3000),
			BlobHashes:    []types.Hash{{0x09}},
		},
		Result: &substate.Result{
			Status: 1,
			Bloom:  types.Bloom{0x0A},
			Logs:   []*types.Log{{Address: types.Address{0x0B}, Topics: []types.Hash{{0x0C}}, Data: []byte{0x0D}}},
		},
		Block:       1,
		Transaction: 1,
	}

	// when
	data, err := Encode(input, 0, 0)

	// then
	assert.Nil(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, "5efc7953a42150436187e9db964d96b8", bytesToMD5(data))
}
