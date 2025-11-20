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
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/testutil"
	"go.uber.org/mock/gomock"
)

func TestBlockHashDb_GetBlockHash(t *testing.T) {
	ctrl := gomock.NewController(t)

	blk := 10
	tests := []struct {
		name      string
		expect    func(mockAidaDb *MockDbAdapter)
		wantHash  types.Hash
		wantError bool
	}{
		{
			name: "GetBlockHash_OK",
			expect: func(mockAidaDb *MockDbAdapter) {
				mockAidaDb.EXPECT().Get(BlockHashDBKey(uint64(blk)), nil).Return(types.Hash{0x11}.Bytes(), nil)
			},
			wantHash:  types.Hash{0x11},
			wantError: false,
		},
		{
			name: "GetBlockHash_NilHash",
			expect: func(mockAidaDb *MockDbAdapter) {
				mockAidaDb.EXPECT().Get(BlockHashDBKey(uint64(blk)), nil).Return(nil, nil)
			},
			wantHash:  types.Hash{},
			wantError: false,
		},
		{
			name: "GetBlockHash_DBError",
			expect: func(mockDb *MockDbAdapter) {
				mockDb.EXPECT().Get(BlockHashDBKey(uint64(blk)), nil).Return(nil, errors.New("db error"))
			},
			wantHash:  types.Hash{},
			wantError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockAdapter := NewMockDbAdapter(ctrl)
			mockDb := NewMockCodeDB(ctrl)
			mockDb.EXPECT().GetBackend().Return(mockAdapter)
			test.expect(mockAdapter)
			bhdb := MakeDefaultBlockHashDBFromBaseDB(mockDb)
			got, err := bhdb.GetBlockHash(blk)
			if test.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, got, test.wantHash)
		})
	}
}

func TestBlockHashDb_GetFirstBlockHash(t *testing.T) {
	testDb := generateTestBlockHashDb(t)
	defer func() {
		err := testDb.Close()
		assert.NoError(t, err, "error closing test database")
	}()

	keyFirst := fmt.Sprintf("0x%x", 1+rand.Intn(1000))
	err := testDb.PutBlockHashString(keyFirst, "0x1234")
	assert.NoError(t, err, "error saving blockHash "+keyFirst)
	keyLast := fmt.Sprintf("0x%x", 1000+rand.Intn(1000))
	err = testDb.PutBlockHashString(keyLast, "0x1234")
	assert.NoError(t, err, "error saving blockHash "+keyLast)

	output, err := testDb.GetFirstKey()
	assert.NoError(t, err)

	assert.Equal(t, keyFirst, "0x"+strconv.FormatUint(output, 16))
}

func TestBlockHashDb_GetFirstBlockHashMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := NewMockDbAdapter(ctrl)
	kValues := &testutil.KeyValue{}
	kValues.PutU(BlockHashDBKey(3), []byte("value"))
	iterValues := iterator.NewArrayIterator(kValues)
	mockAdapter.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterValues)

	mockDb := NewMockCodeDB(ctrl)
	mockDb.EXPECT().GetBackend().Return(mockAdapter)
	blockHash := MakeDefaultBlockHashDBFromBaseDB(mockDb)
	output, err := blockHash.GetFirstKey()
	assert.Equal(t, uint64(0x3), output)
	assert.NoError(t, err)
}

func TestBlockHashDb_GetFirstBlockHashMock_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := NewMockDbAdapter(ctrl)
	kValues := &testutil.KeyValue{}
	iterValues := iterator.NewArrayIterator(kValues)
	mockAdapter.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterValues)

	mockDb := NewMockCodeDB(ctrl)
	mockDb.EXPECT().GetBackend().Return(mockAdapter)
	blockHash := MakeDefaultBlockHashDBFromBaseDB(mockDb)
	output, err := blockHash.GetFirstKey()
	assert.Equal(t, uint64(0x0), output)
	assert.Equal(t, "no blockHash found", err.Error())
}

func TestBlockHashDb_GetLastBlockHash(t *testing.T) {
	testDb := generateTestBlockHashDb(t)
	defer func() {
		err := testDb.Close()
		assert.NoError(t, err, "error closing test database")
	}()

	keyFirst := fmt.Sprintf("0x%x", 1+rand.Intn(1000))
	err := testDb.PutBlockHashString(keyFirst, "0x1234")
	assert.NoError(t, err, "error saving blockHash "+keyFirst)
	keyLast := fmt.Sprintf("0x%x", 1000+rand.Intn(1000))
	err = testDb.PutBlockHashString(keyLast, "0x1234")
	assert.NoError(t, err, "error saving blockHash "+keyLast)

	output, err := testDb.GetLastKey()
	assert.NoError(t, err)
	assert.Equal(t, keyLast, "0x"+strconv.FormatUint(output, 16))
}

func generateTestBlockHashDb(t *testing.T) BlockHashDB {
	tmpDir := t.TempDir() + "/blockHashDb"
	database, err := NewBlockHashDB(tmpDir, nil, nil, nil)
	if err != nil {
		t.Fatalf("error opening blockHash leveldb %s: %v", tmpDir, err)
	}

	return database
}

func TestBlockHashDb_DecodeBlockHashDBKey_Errors(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		want    uint64
		wantErr string
	}{
		{"valid key", append([]byte(BlockHashPrefix), binary.BigEndian.AppendUint64(nil, uint64(2))...), 2, ""},
		{"invalid key", []byte("invalidkey"), 0, "invalid prefix of blockHash key"},
		{"invalid key", []byte("shrt"), 0, "invalid length of blockHash key, expected at least 10, got 4"},
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

func TestBlockHashDb_GetLastBlockHash_EmptyDb(t *testing.T) {
	tmpDir := t.TempDir() + "/blockHashDb"
	database, err := NewBlockHashDB(tmpDir, nil, nil, nil)
	if err != nil {
		t.Fatalf("error opening blockHash leveldb %s: %v", tmpDir, err)
	}
	defer func() {
		err := database.Close()
		assert.NoError(t, err, "error closing test database")
	}()

	output, err := database.GetLastKey()
	if err == nil {
		t.Fatalf("expected error when getting last blockHash from empty db, but got nil")
	}
	assert.Equal(t, "no blockHash found", err.Error())
	assert.Equal(t, uint64(0), output)
}

func TestBlockHashDb_LastBlockHash_InvalidKey(t *testing.T) {
	tmpDir := t.TempDir() + "/blockHashDb"
	database, err := NewBlockHashDB(tmpDir, nil, nil, nil)
	if err != nil {
		t.Fatalf("error opening blockHash leveldb %s: %v", tmpDir, err)
	}

	// Save an invalid blockHash key
	err = database.Put([]byte(BlockHashPrefix+"inv"), []byte("someValue"))
	if err != nil {
		t.Fatalf("error saving invalid blockHash key: %v", err)
	}

	defer func() {
		err := database.Close()
		assert.NoError(t, err, "error closing test database")
	}()

	output, err := database.GetLastKey()
	if err == nil {
		t.Fatalf("expected error when getting last blockHash with invalid key, but got nil")
	}
	assert.Equal(t, "invalid length of blockHash key, expected at least 10, got 5", err.Error())
	assert.Equal(t, uint64(0), output)
}

func TestBlockHashDBConstructors(t *testing.T) {
	tmpDir := t.TempDir()

	// Test NewDefaultBlockHashDB
	db1, err := NewDefaultBlockHashDB(tmpDir + "/db1")
	if err != nil || db1 == nil {
		t.Errorf("NewDefaultBlockHashDB failed: %v", err)
	}

	// Test NewBlockHashDB with custom options
	optOptions := &opt.Options{ErrorIfMissing: false}
	wo := &opt.WriteOptions{Sync: true}
	ro := &opt.ReadOptions{DontFillCache: true}
	db2, err := NewBlockHashDB(tmpDir+"/db2", optOptions, wo, ro)
	if err != nil || db2 == nil {
		t.Errorf("NewBlockHashDB failed: %v", err)
	}

	// Test MakeDefaultBlockHashDB
	ldb3, err := leveldb.OpenFile(tmpDir+"/db3", nil)
	if err != nil {
		t.Fatalf("leveldb.OpenFile failed: %v", err)
	}
	db3 := MakeDefaultBlockHashDB(ldb3)
	if db3 == nil {
		t.Errorf("MakeDefaultBlockHashDB failed")
	}

	// Test MakeDefaultBlockHashDBFromBaseDB
	db4 := MakeDefaultBlockHashDBFromBaseDB(db3)
	if db4 == nil {
		t.Errorf("MakeDefaultBlockHashDBFromBaseDB failed")
	}

	assert.NoError(t, ldb3.Close(), "error closing leveldb")

	// Test NewReadOnlyBlockHashDB
	db5, err := NewReadOnlyBlockHashDB(tmpDir + "/db3")
	if err != nil || db5 == nil {
		t.Errorf("NewReadOnlyBlockHashDB failed: %v", err)
	}

	ldb6, err := leveldb.OpenFile(tmpDir+"/db6", nil)
	if err != nil {
		t.Fatalf("leveldb.OpenFile failed: %v", err)
	}
	// Test MakeBlockHashDB
	wo2 := &opt.WriteOptions{Sync: false}
	ro2 := &opt.ReadOptions{DontFillCache: false}
	db6 := MakeBlockHashDB(ldb6, wo2, ro2)
	if db6 == nil {
		t.Errorf("MakeBlockHashDB failed")
	}
}
