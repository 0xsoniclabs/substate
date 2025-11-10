// Copyright 2025 Sonic Labs
// This file is part of Aida Testing Infrastructure for Sonic
//
// Aida is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Aida is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Aida. If not, see <http://www.gnu.org/licenses/>.

package db

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/mock/gomock"
)

func TestStateHash_KeyToUint64(t *testing.T) {
	type args struct {
		hexBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{"testZeroConvert", args{[]byte(StateRootHashPrefix + "0x0")}, 0, false},
		{"testOneConvert", args{[]byte(StateRootHashPrefix + "0x1")}, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StateHashKeyToUint64(tt.args.hexBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("StateHashKeyToUint64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StateHashKeyToUint64() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStateHash_GetStateRootHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// case success
	mockDb := NewMockBaseDB(ctrl)
	mockDb.EXPECT().Get(gomock.Any()).Return([]byte("abcdefghijabcdefghijabcdefghij32"), nil)
	stateHash := MakeHashProvider(mockDb)
	hash, err := stateHash.GetStateRootHash(1234)
	assert.NoError(t, err)
	assert.Equal(t, "0x6162636465666768696a6162636465666768696a6162636465666768696a3332", hash.String())

	// case error
	mockDb = NewMockBaseDB(ctrl)
	mockDb.EXPECT().Get(gomock.Any()).Return(nil, leveldb.ErrNotFound)
	stateHash = MakeHashProvider(mockDb)
	hash, err = stateHash.GetStateRootHash(1234)
	assert.Equal(t, leveldb.ErrNotFound, err)
	assert.Equal(t, types.Hash{}, hash)

	// case empty
	mockDb = NewMockBaseDB(ctrl)
	mockDb.EXPECT().Get(gomock.Any()).Return(nil, nil)
	stateHash = MakeHashProvider(mockDb)
	hash, err = stateHash.GetStateRootHash(1234)
	assert.NoError(t, err)
	assert.Equal(t, types.Hash{}, hash)
}

func TestStateHash_SaveStateRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// case success
	mockDb := NewMockBaseDB(ctrl)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil)
	err := SaveStateRoot(mockDb, "0x1234", "0x5678")
	assert.NoError(t, err)

	// case error
	mockDb = NewMockBaseDB(ctrl)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(leveldb.ErrNotFound)
	err = SaveStateRoot(mockDb, "0x1234", "0x5678")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "leveldb: not found")
}

func TestStateHash_StateHashKeyToUint64(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// case success
	output, err := StateHashKeyToUint64([]byte("dbh0x1234"))
	assert.NoError(t, err)
	assert.Equal(t, uint64(0x1234), output)

	// case error
	output, err = StateHashKeyToUint64([]byte("ggggggg"))
	assert.Equal(t, uint64(0), output)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid syntax")
}

func TestStateHash_retrieveStateRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// case success
	client := NewMockIRpcClient(ctrl)
	client.EXPECT().Call(gomock.Any(), "eth_getBlockByNumber", "0x1234", false).Return(nil)
	output, err := GetBlockByNumber(client, "0x1234")
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}(nil), output)

	// case error
	mockErr := errors.New("error")
	client = NewMockIRpcClient(ctrl)
	client.EXPECT().Call(gomock.Any(), "eth_getBlockByNumber", "0x1234", false).Return(mockErr)
	output, err = GetBlockByNumber(client, "0x1234")
	assert.Error(t, err)
	assert.Nil(t, output)
}

func TestStateHash_GetFirstStateHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockBaseDB(ctrl)
	output, err := GetFirstStateHash(mockDb)
	assert.Equal(t, uint64(0x0), output)
	assert.Error(t, err)

}

func TestStateHash_GetLastStateHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockBaseDB(ctrl)
	output, err := GetLastStateHash(mockDb)
	assert.Equal(t, uint64(0x0), output)
	assert.Error(t, err)
}

func TestStateHashProvider_GetStateRootHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	blk := 10
	tests := []struct {
		name   string
		expect func(mockAidaDb *MockBaseDB)
		want   types.Hash
	}{
		{
			name: "GetStatRootHash_OK",
			expect: func(mockAidaDb *MockBaseDB) {
				hex := strconv.FormatUint(uint64(blk), 16)
				mockAidaDb.EXPECT().Get([]byte(StateRootHashPrefix+"0x"+hex)).Return(types.Hash{0x11}.Bytes(), nil)
			},
			want: types.Hash{0x11},
		},
		{
			name: "GetStatRootHash_NilHash",
			expect: func(mockAidaDb *MockBaseDB) {
				hex := strconv.FormatUint(uint64(blk), 16)
				mockAidaDb.EXPECT().Get([]byte(StateRootHashPrefix+"0x"+hex)).Return(nil, nil)
			},
			want: types.Hash{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockAidaDb := NewMockBaseDB(ctrl)
			hp := MakeHashProvider(mockAidaDb)
			test.expect(mockAidaDb)
			got, err := hp.GetStateRootHash(blk)
			assert.NoError(t, err)
			assert.Equal(t, got, test.want)
		})
	}
}

func TestStateHashProvider_GetBlockHash(t *testing.T) {
	ctrl := gomock.NewController(t)

	blk := 10
	tests := []struct {
		name      string
		expect    func(mockAidaDb *MockBaseDB)
		wantHash  types.Hash
		wantError bool
	}{
		{
			name: "GetBlockHash_OK",
			expect: func(mockAidaDb *MockBaseDB) {
				mockAidaDb.EXPECT().Get(BlockHashDBKey(uint64(blk))).Return(types.Hash{0x11}.Bytes(), nil)
			},
			wantHash:  types.Hash{0x11},
			wantError: false,
		},
		{
			name: "GetBlockHash_NilHash",
			expect: func(mockAidaDb *MockBaseDB) {
				mockAidaDb.EXPECT().Get(BlockHashDBKey(uint64(blk))).Return(nil, nil)
			},
			wantHash:  types.Hash{},
			wantError: false,
		},
		{
			name: "GetBlockHash_DBError",
			expect: func(mockAidaDb *MockBaseDB) {
				mockAidaDb.EXPECT().Get(BlockHashDBKey(uint64(blk))).Return(nil, errors.New("db error"))
			},
			wantHash:  types.Hash{},
			wantError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockAidaDb := NewMockBaseDB(ctrl)
			hp := MakeHashProvider(mockAidaDb)
			test.expect(mockAidaDb)
			got, err := hp.GetBlockHash(blk)
			if test.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, got, test.wantHash)
		})
	}
}

func TestStateHash_GetFirstBlockHash(t *testing.T) {
	testDb := generateTestBlockHashDb(t)
	defer func() {
		err := testDb.Close()
		assert.NoError(t, err, "error closing test database")
	}()

	keyFirst := fmt.Sprintf("0x%x", 1+rand.Intn(1000))
	err := SaveBlockHash(testDb, keyFirst, "0x1234")
	assert.NoError(t, err, "error saving state root "+keyFirst)
	keyLast := fmt.Sprintf("0x%x", 1000+rand.Intn(1000))
	err = SaveBlockHash(testDb, keyLast, "0x1234")
	assert.NoError(t, err, "error saving state root "+keyLast)

	output, err := GetFirstBlockHash(testDb)
	assert.NoError(t, err)

	assert.Equal(t, keyFirst, "0x"+strconv.FormatUint(output, 16))
}

func TestStateHash_GetLastBlockHash(t *testing.T) {
	testDb := generateTestBlockHashDb(t)
	defer func() {
		err := testDb.Close()
		assert.NoError(t, err, "error closing test database")
	}()

	keyFirst := fmt.Sprintf("0x%x", 1+rand.Intn(1000))
	err := SaveBlockHash(testDb, keyFirst, "0x1234")
	assert.NoError(t, err, "error saving state root "+keyFirst)
	keyLast := fmt.Sprintf("0x%x", 1000+rand.Intn(1000))
	err = SaveBlockHash(testDb, keyLast, "0x1234")
	assert.NoError(t, err, "error saving state root "+keyLast)

	output, err := GetLastBlockHash(testDb)
	assert.NoError(t, err)
	assert.Equal(t, keyLast, "0x"+strconv.FormatUint(output, 16))
}

func generateTestBlockHashDb(t *testing.T) BaseDB {
	tmpDir := t.TempDir() + "/blockHashDb"
	database, err := NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatalf("error opening stateHash leveldb %s: %v", tmpDir, err)
	}

	return database
}

func TestDecodeBlockHashDBKey_Errors(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		want    uint64
		wantErr string
	}{
		{"valid key", append([]byte(BlockHashPrefix), binary.BigEndian.AppendUint64(nil, uint64(2))...), 2, ""},
		{"invalid key", []byte("invalidkey"), 0, "invalid prefix of block hash key"},
		{"invalid key", []byte("shrt"), 0, "invalid length of block hash key, expected at least 10, got 4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeBlockHashDBKey(tt.key)

			if err != nil {
				if err.Error() != tt.wantErr {
					t.Errorf("DecodeBlockHashDBKey() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			} else if tt.wantErr != "" {
				t.Errorf("DecodeBlockHashDBKey() expected error %v, got nil", tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("DecodeBlockHashDBKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLastBlockHash_EmptyDb(t *testing.T) {
	tmpDir := t.TempDir() + "/blockHashDb"
	database, err := NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatalf("error opening stateHash leveldb %s: %v", tmpDir, err)
	}
	defer func() {
		err := database.Close()
		assert.NoError(t, err, "error closing test database")
	}()

	output, err := GetLastBlockHash(database)
	if err == nil {
		t.Fatalf("expected error when getting last block hash from empty db, but got nil")
	}
	assert.Equal(t, "no block hash found", err.Error())
	assert.Equal(t, uint64(0), output)
}

func TestGetLastBlockHash_InvalidKey(t *testing.T) {
	tmpDir := t.TempDir() + "/blockHashDb"
	database, err := NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatalf("error opening stateHash leveldb %s: %v", tmpDir, err)
	}

	// Save an invalid block hash key
	err = database.Put([]byte(BlockHashPrefix+"inv"), []byte("someValue"))
	if err != nil {
		t.Fatalf("error saving invalid block hash key: %v", err)
	}

	defer func() {
		err := database.Close()
		assert.NoError(t, err, "error closing test database")
	}()

	output, err := GetLastBlockHash(database)
	if err == nil {
		t.Fatalf("expected error when getting last block hash with invalid key, but got nil")
	}
	assert.Equal(t, "invalid length of block hash key, expected at least 10, got 5", err.Error())
	assert.Equal(t, uint64(0), output)
}
