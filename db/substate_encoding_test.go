package db

import (
	"math/big"
	"strings"
	"testing"

	"github.com/holiman/uint256"

	"github.com/0xsoniclabs/substate/protobuf"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"

	pb "github.com/0xsoniclabs/substate/protobuf"
)

type encTest struct {
	bytes []byte
	blk   uint64
	tx    int
}

var (
	blk = getTestSubstate("default").Block
	tx  = getTestSubstate("default").Transaction

	simplePb, _ = pb.Encode(getTestSubstate("protobuf"), blk, tx)
	testPb      = encTest{bytes: simplePb, blk: blk, tx: tx}

	supportedEncoding = map[string]encTest{
		"protobuf": testPb,
	}
)

func TestSubstateEncoding_UnsupportedEncodingThrowsError(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := newSubstateDB(path, nil, nil, nil)
	if err != nil {
		t.Errorf("cannot open db; %v", err)
	}

	err = db.SetSubstateEncoding("EncodingNotSupported")
	if err == nil || !strings.Contains(err.Error(), "encoding not supported") {
		t.Error("encoding not supported, but no error")
	}
}

func TestSubstateEncoding_TestDb(t *testing.T) {
	for encoding, et := range supportedEncoding {
		path := t.TempDir() + "test-db-" + encoding
		db, err := newSubstateDB(path, nil, nil, nil)
		if err != nil {
			t.Fatalf("cannot open db; %v", err)
		}

		ts := getTestSubstate(encoding)
		err = db.SetSubstateEncoding(encoding)
		if err != nil {
			t.Fatal(err)
		}

		ss, err := db.decodeToSubstate(et.bytes, et.blk, et.tx)
		if err != nil {
			t.Fatal(err)
		}

		err = addCustomSubstate(db, et.blk, ss)
		if err != nil {
			t.Fatal(err)
		}

		err = testSubstateDB_GetSubstate(db, *ts)
		if err != nil {
			t.Fatalf("getSubstate check failed: encoding: %s; err: %v", encoding, err)
		}
	}
}

func TestSubstateEncoding_TestIterator(t *testing.T) {
	for encoding, et := range supportedEncoding {
		path := t.TempDir() + "test-db-" + encoding
		db, err := newSubstateDB(path, nil, nil, nil)
		if err != nil {
			t.Errorf("cannot open db; %v", err)
		}

		err = db.SetSubstateEncoding(encoding)
		if err != nil {
			t.Error(err)
		}

		ss, err := db.decodeToSubstate(et.bytes, et.blk, et.tx)
		if err != nil {
			t.Error(err)
		}

		err = addCustomSubstate(db, et.blk, ss)
		if err != nil {
			t.Error(err)
		}

		testSubstatorIterator_Value(db, t)
	}
}

func getSubstate() *substate.Substate {
	return &substate.Substate{
		InputSubstate: substate.WorldState{
			types.Address{0x01}: &substate.Account{
				Nonce:   1,
				Balance: uint256.NewInt(1000),
				Storage: map[types.Hash]types.Hash{
					{0x01}: {0x02},
				},
			},
		},
		OutputSubstate: substate.WorldState{
			types.Address{0x04}: &substate.Account{
				Nonce:   1,
				Balance: uint256.NewInt(2000),
				Storage: map[types.Hash]types.Hash{
					{0xCD}: {0xAB},
				},
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
			AccessList:     []types.AccessTuple{},
			GasFeeCap:      big.NewInt(100),
			GasTipCap:      big.NewInt(100),
			BlobGasFeeCap:  big.NewInt(3000),
			BlobHashes:     nil,
		},
		Result: &substate.Result{
			Status: 1,
			Bloom:  types.Bloom{0x0A},
			Logs:   []*types.Log{{Address: types.Address{0x0B}, Topics: []types.Hash{{0x0C}}, Data: []byte{0x0D}}},
		},
		Block:       1,
		Transaction: 1,
	}
}

func TestSubstateDB_SetSubstateEncodingSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)

	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}

	err := db.SetSubstateEncoding("")
	assert.Nil(t, err)
}

func TestSubstateDB_SetSubstateEncodingFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)

	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}

	err := db.SetSubstateEncoding("xyz")
	assert.NotNil(t, err)
}

func TestSubstateDB_GetSubstateEncodingNilSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)

	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}

	value := db.GetSubstateEncoding()
	assert.Equal(t, "", value)
}

func TestSubstateDB_GetSubstateEncodingSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)

	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}
	err := db.SetSubstateEncoding("protobuf")
	assert.Nil(t, err)

	value := db.GetSubstateEncoding()
	assert.Equal(t, "protobuf", value)
}

func TestSubstateDB_DecodeToSubstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	mockDb.EXPECT().GetCode(gomock.Any()).Return(nil, nil).AnyTimes()
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}
	input := getSubstate()
	encoded, err := protobuf.Encode(input, 1, 1)
	assert.Nil(t, err)
	err = db.SetSubstateEncoding("protobuf")
	assert.Nil(t, err)

	value, err := db.decodeToSubstate(encoded, 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, input, value)
}

func TestSubstateDB_EncodeToSubstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockCodeDB(ctrl)
	mockDb.EXPECT().GetCode(gomock.Any()).Return(nil, nil).AnyTimes()
	db := &substateDB{
		CodeDB:   mockDb,
		encoding: nil,
	}
	input := getSubstate()
	encoded, err := protobuf.Encode(input, 1, 1)
	assert.Nil(t, err)
	err = db.SetSubstateEncoding("protobuf")
	assert.Nil(t, err)

	value, err := db.encodeSubstate(input, 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, encoded, value)
}

func TestDecodeProtobuf_Success(t *testing.T) {
	value, err := decodeProtobuf([]byte{1, 2, 3}, nil, 1, 1)
	assert.NotNil(t, err)
	assert.Nil(t, value)

}
