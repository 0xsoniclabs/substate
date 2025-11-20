package db

import (
	"strconv"
	"testing"

	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"go.uber.org/mock/gomock"
)

func TestStateHashDb_KeyToUint64(t *testing.T) {
	type args struct {
		hexBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{"testZeroConvert", args{[]byte(StateHashPrefix + "0x0")}, 0, false},
		{"testOneConvert", args{[]byte(StateHashPrefix + "0x1")}, 1, false},
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

func TestStateHashDb_GetStateRootHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// case success
	mockDbAdapter := NewMockDbAdapter(ctrl)
	mockDb := NewMockCodeDB(ctrl)
	mockDb.EXPECT().GetBackend().Return(mockDbAdapter)
	stateHash := MakeDefaultStateHashDBFromBaseDB(mockDb)
	mockDbAdapter.EXPECT().Get(gomock.Any(), gomock.Any()).Return([]byte("abcdefghijabcdefghijabcdefghij32"), nil)
	hash, err := stateHash.GetStateHash(1234)
	assert.NoError(t, err)
	assert.Equal(t, "0x6162636465666768696a6162636465666768696a6162636465666768696a3332", hash.String())

	// case error
	mockDbAdapter = NewMockDbAdapter(ctrl)
	mockDb = NewMockCodeDB(ctrl)
	mockDb.EXPECT().GetBackend().Return(mockDbAdapter)
	mockDbAdapter.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, leveldb.ErrNotFound)
	stateHash = MakeDefaultStateHashDBFromBaseDB(mockDb)
	hash, err = stateHash.GetStateHash(1234)
	assert.Equal(t, leveldb.ErrNotFound, err)
	assert.Equal(t, types.Hash{}, hash)

	// case empty
	mockDbAdapter = NewMockDbAdapter(ctrl)
	mockDb = NewMockCodeDB(ctrl)
	mockDb.EXPECT().GetBackend().Return(mockDbAdapter)
	mockDbAdapter.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, nil)
	stateHash = MakeDefaultStateHashDBFromBaseDB(mockDb)
	hash, err = stateHash.GetStateHash(1234)
	assert.NoError(t, err)
	assert.Equal(t, types.Hash{}, hash)
}

func TestStateHashDb_PutStateHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// case success
	mockDbAdapter := NewMockDbAdapter(ctrl)
	mockDb := NewMockCodeDB(ctrl)
	mockDb.EXPECT().GetBackend().Return(mockDbAdapter)
	mockDbAdapter.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	stateHash := MakeDefaultStateHashDBFromBaseDB(mockDb)
	err := stateHash.PutStateHash(123, []byte("345"))
	assert.NoError(t, err)

	// case error
	mockDb = NewMockCodeDB(ctrl)
	mockDb.EXPECT().GetBackend().Return(mockDbAdapter)
	mockDbAdapter.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).Return(leveldb.ErrNotFound)
	stateHash = MakeDefaultStateHashDBFromBaseDB(mockDb)
	err = stateHash.PutStateHash(123, []byte("5678"))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "leveldb: not found")
}

func TestStateHashDb_PutStateHashString(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// case success
	mockDbAdapter := NewMockDbAdapter(ctrl)
	mockDb := NewMockCodeDB(ctrl)
	mockDb.EXPECT().GetBackend().Return(mockDbAdapter)
	mockDbAdapter.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	stateHash := MakeDefaultStateHashDBFromBaseDB(mockDb)
	err := stateHash.PutStateHashString("0x1234", "0x5678")
	assert.NoError(t, err)

	// case error
	mockDb = NewMockCodeDB(ctrl)
	mockDb.EXPECT().GetBackend().Return(mockDbAdapter)
	mockDbAdapter.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).Return(leveldb.ErrNotFound)
	stateHash = MakeDefaultStateHashDBFromBaseDB(mockDb)
	err = stateHash.PutStateHashString("0x1234", "0x5678")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "leveldb: not found")
}

func TestStateHashDb_StateHashKeyToUint64(t *testing.T) {
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

func TestStateHashDb_GetFirstStateHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := NewMockDbAdapter(ctrl)
	mockDb := NewMockCodeDB(ctrl)
	mockDb.EXPECT().GetBackend().Return(mockAdapter)
	stateHash := MakeDefaultStateHashDBFromBaseDB(mockDb)
	output, err := stateHash.GetFirstKey()
	assert.Equal(t, uint64(0x0), output)
	assert.Error(t, err)
}

func TestStateHashDb_GetLastStateHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDbAdapter := NewMockDbAdapter(ctrl)
	mockDb := NewMockCodeDB(ctrl)
	mockDb.EXPECT().GetBackend().Return(mockDbAdapter)
	//TODO fix after stateHash gets converted
	//mockDbAdapter.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, leveldb.ErrNotFound)
	stateHash := MakeDefaultStateHashDBFromBaseDB(mockDb)
	output, err := stateHash.GetLastKey()
	assert.Equal(t, uint64(0x0), output)
	assert.Error(t, err)
}

func TestStateHashDb_GetStateRootHash_IntegrationTests(t *testing.T) {
	ctrl := gomock.NewController(t)
	blk := 10
	tests := []struct {
		name   string
		expect func(mockAidaDb *MockDbAdapter)
		want   types.Hash
	}{
		{
			name: "GetStatRootHash_OK",
			expect: func(mockAidaDb *MockDbAdapter) {
				hex := strconv.FormatUint(uint64(blk), 16)
				mockAidaDb.EXPECT().Get([]byte(StateHashPrefix+"0x"+hex), nil).Return(types.Hash{0x11}.Bytes(), nil)
			},
			want: types.Hash{0x11},
		},
		{
			name: "GetStatRootHash_NilHash",
			expect: func(mockAidaDb *MockDbAdapter) {
				hex := strconv.FormatUint(uint64(blk), 16)
				mockAidaDb.EXPECT().Get([]byte(StateHashPrefix+"0x"+hex), nil).Return(nil, nil)
			},
			want: types.Hash{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			mockAdapter := NewMockDbAdapter(ctrl)
			mockDb := NewMockCodeDB(ctrl)
			mockDb.EXPECT().GetBackend().Return(mockAdapter)
			test.expect(mockAdapter)
			shdb := MakeDefaultStateHashDBFromBaseDB(mockDb)
			got, err := shdb.GetStateHash(blk)
			assert.NoError(t, err)
			assert.Equal(t, got, test.want)
		})
	}
}

func TestStateHashDBConstructors(t *testing.T) {
	tmpDir := t.TempDir()

	// Test NewDefaultStateHashDB
	db1, err := NewDefaultStateHashDB(tmpDir + "/db1")
	if err != nil || db1 == nil {
		t.Errorf("NewDefaultStateHashDB failed: %v", err)
	}

	// Test NewStateHashDB with custom options
	optOptions := &opt.Options{ErrorIfMissing: false}
	wo := &opt.WriteOptions{Sync: true}
	ro := &opt.ReadOptions{DontFillCache: true}
	db2, err := NewStateHashDB(tmpDir+"/db2", optOptions, wo, ro)
	if err != nil || db2 == nil {
		t.Errorf("NewStateHashDB failed: %v", err)
	}

	// Test MakeDefaultStateHashDB
	ldb3, err := leveldb.OpenFile(tmpDir+"/db3", nil)
	if err != nil {
		t.Fatalf("leveldb.OpenFile failed: %v", err)
	}
	db3 := MakeDefaultStateHashDB(ldb3)
	if db3 == nil {
		t.Errorf("MakeDefaultStateHashDB failed")
	}

	// Test MakeDefaultStateHashDBFromBaseDB
	db4 := MakeDefaultStateHashDBFromBaseDB(db3)
	if db4 == nil {
		t.Errorf("MakeDefaultStateHashDBFromBaseDB failed")
	}

	assert.NoError(t, ldb3.Close(), "error closing leveldb")

	// Test NewReadOnlyStateHashDB
	db5, err := NewReadOnlyStateHashDB(tmpDir + "/db3")
	if err != nil || db5 == nil {
		t.Errorf("NewReadOnlyStateHashDB failed: %v", err)
	}

	ldb6, err := leveldb.OpenFile(tmpDir+"/db6", nil)
	if err != nil {
		t.Fatalf("leveldb.OpenFile failed: %v", err)
	}
	// Test MakeStateHashDB
	wo2 := &opt.WriteOptions{Sync: false}
	ro2 := &opt.ReadOptions{DontFillCache: false}
	db6 := MakeStateHashDB(ldb6, wo2, ro2)
	if db6 == nil {
		t.Errorf("MakeStateHashDB failed")
	}
	assert.NoError(t, ldb6.Close(), "error closing leveldb")
}
