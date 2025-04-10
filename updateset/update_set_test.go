package updateset

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
)

func TestUpdateSet_ToWorldStateRLP(t *testing.T) {
	// Create test accounts
	acc1 := substate.NewAccount(1, new(uint256.Int).SetUint64(100), []byte{1, 2, 3})
	acc1.Storage = map[types.Hash]types.Hash{
		{1}: {2},
	}

	acc2 := substate.NewAccount(2, new(uint256.Int).SetUint64(200), []byte{4, 5, 6})

	// Create world state
	ws := substate.WorldState{
		types.Address{1}: acc1,
		types.Address{2}: acc2,
	}

	// Create update set
	updateSet := NewUpdateSet(ws, 10)

	// Convert to PB
	pbState := updateSet.ToWorldStatePB()

	// Verify word state length
	assert.Equal(t, len(ws), len(pbState.Alloc))

	// Find the accounts in the RLP representation
	var foundAcc1, foundAcc2 bool
	for _, alloc := range pbState.Alloc {
		acc := alloc.Account
		addr := types.BytesToAddress(alloc.Address)

		if addr == (types.Address{1}) {
			foundAcc1 = true
			assert.Equal(t, uint64(1), acc.GetNonce())
			assert.Equal(t, uint64(100), bytesToUint64(acc.GetBalance()))
			assert.Equal(t, acc1.CodeHash(), types.BytesToHash(acc.GetCodeHash()))
			assert.Equal(t, 1, len(acc.GetStorage()))
		}

		if addr == (types.Address{2}) {
			foundAcc2 = true
			assert.Equal(t, uint64(2), acc.GetNonce())
			assert.Equal(t, uint64(200), bytesToUint64(acc.GetBalance()))
			assert.Equal(t, acc2.CodeHash(), types.BytesToHash(acc.GetCodeHash()))
			assert.Equal(t, 0, len(acc.GetStorage()))
		}
	}

	assert.True(t, foundAcc1)
	assert.True(t, foundAcc2)
}

func TestUpdateSet_Equal(t *testing.T) {
	// Create test accounts
	acc1 := substate.NewAccount(1, new(uint256.Int).SetUint64(100), []byte{1, 2, 3})
	acc2 := substate.NewAccount(2, new(uint256.Int).SetUint64(200), []byte{4, 5, 6})

	// Create first world state
	ws1 := substate.WorldState{
		types.Address{1}: acc1,
		types.Address{2}: acc2,
	}

	// Create identical second world state
	ws2 := substate.WorldState{
		types.Address{1}: acc1.Copy(),
		types.Address{2}: acc2.Copy(),
	}

	// Different world state
	ws3 := substate.WorldState{
		types.Address{1}: acc1.Copy(),
	}

	// Create update sets
	updateSet1 := NewUpdateSet(ws1, 10)
	updateSet2 := NewUpdateSet(ws2, 10)
	updateSet3 := NewUpdateSet(ws1, 11) // Different block number
	updateSet4 := NewUpdateSet(ws3, 10) // Different accounts

	// Set deleted accounts
	updateSet1.DeletedAccounts = []types.Address{{3}}
	updateSet2.DeletedAccounts = []types.Address{{3}}
	updateSet3.DeletedAccounts = []types.Address{{3}}
	updateSet4.DeletedAccounts = []types.Address{{3}}

	// Test equality
	assert.True(t, updateSet1.Equal(updateSet2))
	assert.False(t, updateSet1.Equal(updateSet3))
	assert.False(t, updateSet1.Equal(updateSet4))

	// Test self equality
	assert.True(t, updateSet1.Equal(updateSet1))

	// Test with different deleted accounts
	updateSet5 := NewUpdateSet(ws1, 10)
	updateSet5.DeletedAccounts = []types.Address{{4}}
	assert.False(t, updateSet1.Equal(updateSet5))
}

func TestUpdateSetRLP_ToWorldStateSuccess(t *testing.T) {
	// Create test accounts
	acc1 := substate.NewAccount(1, new(uint256.Int).SetUint64(100), []byte{1, 2, 3})
	acc1.Storage = map[types.Hash]types.Hash{
		{1}: {2},
	}

	acc2 := substate.NewAccount(2, new(uint256.Int).SetUint64(200), []byte{4, 5, 6})

	// Create world state
	ws := substate.WorldState{
		types.Address{1}: acc1,
		types.Address{2}: acc2,
	}

	// Create update set
	updateSet := NewUpdateSet(ws, 10)
	updateSet.DeletedAccounts = []types.Address{{3}}

	// Create RLP version
	rlpUpdateSet := NewUpdateSetRLP(updateSet, updateSet.DeletedAccounts)

	// Mock getCodeFunc
	getCodeFunc := func(codeHash types.Hash) ([]byte, error) {
		if codeHash == acc1.CodeHash() {
			return []byte{1, 2, 3}, nil
		}
		if codeHash == acc2.CodeHash() {
			return []byte{4, 5, 6}, nil
		}
		return nil, nil
	}

	// Convert back to world state
	newUpdateSet, err := rlpUpdateSet.ToWorldState(getCodeFunc, 10)
	assert.NoError(t, err)

	// Verify the conversion
	assert.Equal(t, uint64(10), newUpdateSet.Block)
	assert.Equal(t, 2, len(newUpdateSet.WorldState))

	// Check accounts are converted correctly
	for addr, account := range newUpdateSet.WorldState {
		if addr == (types.Address{1}) {
			assert.Equal(t, uint64(1), account.Nonce)
			assert.Equal(t, uint64(100), account.Balance.Uint64())
			assert.Equal(t, []byte{1, 2, 3}, account.Code)
			assert.Equal(t, 1, len(account.Storage))
			assert.Equal(t, types.Hash{2}, account.Storage[types.Hash{1}])
		}

		if addr == (types.Address{2}) {
			assert.Equal(t, uint64(2), account.Nonce)
			assert.Equal(t, uint64(200), account.Balance.Uint64())
			assert.Equal(t, []byte{4, 5, 6}, account.Code)
			assert.Equal(t, 0, len(account.Storage))
		}
	}
}

func TestUpdateSetRLP_ToWorldStateError(t *testing.T) {
	// Create test accounts
	acc1 := substate.NewAccount(1, new(uint256.Int).SetUint64(100), []byte{1, 2, 3})
	acc1.Storage = map[types.Hash]types.Hash{
		{1}: {2},
	}

	acc2 := substate.NewAccount(2, new(uint256.Int).SetUint64(200), []byte{4, 5, 6})

	// Create world state
	ws := substate.WorldState{
		types.Address{1}: acc1,
		types.Address{2}: acc2,
	}

	// Create update set
	updateSet := NewUpdateSet(ws, 10)
	updateSet.DeletedAccounts = []types.Address{{3}}

	// Create RLP version
	rlpUpdateSet := NewUpdateSetRLP(updateSet, updateSet.DeletedAccounts)

	// Mock getCodeFunc
	getCodeFunc := func(codeHash types.Hash) ([]byte, error) {
		return nil, leveldb.ErrReadOnly
	}

	// Convert back to world state
	newUpdateSet, err := rlpUpdateSet.ToWorldState(getCodeFunc, 10)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "leveldb: read-only mode")
	assert.Nil(t, newUpdateSet)
}

func TestUpdateSetPB_EncodeUpdateSetPBSuccess(t *testing.T) {
	updateSet := NewUpdateSetRLP(&UpdateSet{
		WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
		Block:           0,
		DeletedAccounts: []types.Address{},
	}, []types.Address{{}})
	value, err := EncodeUpdateSetPB(&updateSet)

	assert.NoError(t, err)
	assert.Equal(t, "06753c366fe2f1b1bdc6d67cb3e3698f", bytesToMD5(value))
}

func TestUpdateSetPB_EncodeUpdateSetPBError(t *testing.T) {
	value, err := EncodeUpdateSetPB(&UpdateSetPB{})
	assert.Error(t, err)
	assert.Nil(t, value)
}

func TestUpdateSetPB_DecodeUpdateSetPBSuccess(t *testing.T) {
	expected := NewUpdateSetRLP(&UpdateSet{
		WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
		Block:           0,
		DeletedAccounts: []types.Address{},
	}, []types.Address{{}})

	value, err := DecodeUpdateSetPB([]byte{0xa, 0x41, 0xa, 0x3f, 0xa, 0x14, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x12, 0x27, 0x8, 0x1, 0x12, 0x1, 0x1, 0x2a, 0x20, 0xc5, 0xd2, 0x46, 0x1, 0x86, 0xf7, 0x23, 0x3c, 0x92, 0x7e, 0x7d, 0xb2, 0xdc, 0xc7, 0x3, 0xc0, 0xe5, 0x0, 0xb6, 0x53, 0xca, 0x82, 0x27, 0x3b, 0x7b, 0xfa, 0xd8, 0x4, 0x5d, 0x85, 0xa4, 0x70, 0x12, 0x14, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	assert.NoError(t, err)
	assert.Equal(t, expected.DeletedAccounts, value.DeletedAccounts)
	assert.Equal(t, len(expected.WorldState.Alloc), len(value.WorldState.Alloc))
	assert.Equal(t, expected.WorldState.Alloc[0].Address, value.WorldState.Alloc[0].Address)
	assert.Equal(t, expected.WorldState.Alloc[0].Account.GetNonce(), value.WorldState.Alloc[0].Account.GetNonce())
	assert.Equal(t, expected.WorldState.Alloc[0].Account.GetBalance(), value.WorldState.Alloc[0].Account.GetBalance())
	assert.Equal(t, expected.WorldState.Alloc[0].Account.GetCode(), value.WorldState.Alloc[0].Account.GetCode())
}

func TestUpdateSetPB_DecodeUpdateSetPBError(t *testing.T) {
	value, err := DecodeUpdateSetPB(nil)
	assert.Error(t, err)
	assert.NotNil(t, value)
}

// bytesToUint64 converts a byte slice to a uint64 value.
func bytesToUint64(b []byte) uint64 {
	// Ensure the byte slice has enough bytes
	if len(b) < 8 {
		// Create a new 8-byte slice
		buf := make([]byte, 8)
		// Copy the input bytes to the end of the buffer
		copy(buf[8-len(b):], b)
		b = buf
	}
	return binary.BigEndian.Uint64(b)
}

// bytesToMD5 computes the MD5 hash of the input byte slice and returns it as a hexadecimal string.
func bytesToMD5(data []byte) string {
	hash := md5.Sum(data)
	hashString := hex.EncodeToString(hash[:])
	return hashString
}
