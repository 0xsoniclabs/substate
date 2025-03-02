package substate

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/0xsoniclabs/substate/types"
)

func TestWorldState_Add(t *testing.T) {
	addr1 := types.Address{1}
	addr2 := types.Address{2}
	acc := &Account{
		Nonce:   2,
		Balance: new(big.Int).SetUint64(2),
		Code:    []byte{2},
	}

	worldState := make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
	worldState.Add(addr2, acc.Nonce, acc.Balance, acc.Code)

	if len(worldState) != 2 {
		t.Fatalf("incorrect len after add\n got: %v\n want: %v", len(worldState), 2)
	}

	if !worldState[addr2].Equal(acc) {
		t.Fatalf("incorrect merge of addr1\ngot: %v\n want: %v", worldState[addr2], acc)
	}
}

func TestWorldState_MergeOneAccount(t *testing.T) {
	addr := types.Address{1}

	worldState := make(WorldState).Add(addr, 1, new(big.Int).SetUint64(1), []byte{1})
	worldStateToMerge := make(WorldState).Add(addr, 2, new(big.Int).SetUint64(2), []byte{2})

	worldState.Merge(worldStateToMerge)

	acc := &Account{
		Nonce:   2,
		Balance: new(big.Int).SetUint64(2),
		Code:    []byte{2},
	}

	if !worldState[addr].Equal(acc) {
		t.Fatalf("incorrect merge\ngot: %v\n want: %v", worldState[addr], acc)
	}

}

func TestWorldState_MergeTwoAccounts(t *testing.T) {
	addr1 := types.Address{1}
	addr2 := types.Address{2}

	worldState := make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
	worldStateToMerge := make(WorldState).Add(addr2, 2, new(big.Int).SetUint64(2), []byte{2})

	worldState.Merge(worldStateToMerge)

	want1 := &Account{
		Nonce:   1,
		Balance: new(big.Int).SetUint64(1),
		Code:    []byte{1},
	}

	if len(worldState) != 2 {
		t.Fatalf("incorrect len after merge\n got: %v\n want: %v", len(worldState), 2)
	}

	if !worldState[addr1].Equal(want1) {
		t.Fatalf("incorrect merge of addr1\ngot: %v\n want: %v", worldState[addr1], want1)
	}

	want2 := &Account{
		Nonce:   2,
		Balance: new(big.Int).SetUint64(2),
		Code:    []byte{2},
	}

	if !worldState[addr2].Equal(want2) {
		t.Fatalf("incorrect merge of addr1\ngot: %v\n want: %v", worldState[addr2], want2)
	}

}

func TestWorldState_EstimateIncrementalSize_NewWorldState(t *testing.T) {
	addr1 := types.Address{1}
	addr2 := types.Address{2}

	worldState := make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
	worldStateToEstimate := make(WorldState).Add(addr2, 2, new(big.Int).SetUint64(2), []byte{2})

	want := sizeOfAddress + sizeOfNonce + uint64(len(worldStateToEstimate[addr2].Balance.Bytes())) + sizeOfHash

	// adding new world state without storage keys
	if got := worldState.EstimateIncrementalSize(worldStateToEstimate); got != want {
		t.Fatalf("incorrect estimation\ngot: %v\nwant: %v", got, want)
	}
}

func TestWorldState_EstimateIncrementalSize_SameWorldState(t *testing.T) {
	addr1 := types.Address{1}

	worldState := make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
	worldStateToEstimate := make(WorldState).Add(addr1, 2, new(big.Int).SetUint64(2), []byte{2})

	// since we don't add anything, size should not be increased
	if got := worldState.EstimateIncrementalSize(worldStateToEstimate); got != 0 {
		t.Fatalf("incorrect estimation\ngot: %v\nwant: %v", got, 0)
	}
}

func TestWorldState_EstimateIncrementalSize_AddingStorageHash(t *testing.T) {
	addr1 := types.Address{1}

	worldState := make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
	worldStateToEstimate := make(WorldState).Add(addr1, 2, new(big.Int).SetUint64(2), []byte{2})
	worldStateToEstimate[addr1].Storage[types.Hash{1}] = types.Hash{1}

	// we add one key to already existing account, this size is increased by the sizeOfHash
	if got := worldState.EstimateIncrementalSize(worldStateToEstimate); got != sizeOfHash {
		t.Fatalf("incorrect estimation\ngot: %v\nwant: %v", got, sizeOfHash)
	}
}

// todo diff tests

func TestWorldState_Equal(t *testing.T) {
	worldState := make(WorldState).Add(types.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
	comparedWorldStateEqual := make(WorldState).Add(types.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})

	if !worldState.Equal(comparedWorldStateEqual) {
		t.Fatal("world states are same but equal returned false")
	}
}

func TestWorldState_NotEqual(t *testing.T) {
	worldState := make(WorldState).Add(types.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
	comparedWorldStateEqual := make(WorldState).Add(types.Address{2}, 1, new(big.Int).SetUint64(1), []byte{1})

	if worldState.Equal(comparedWorldStateEqual) {
		t.Fatal("world states are different but equal returned false")
	}
}

func TestWorldState_NotEqual_DifferentLen(t *testing.T) {
	worldState := make(WorldState).Add(types.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
	comparedWorldStateEqual := make(WorldState).Add(types.Address{2}, 1, new(big.Int).SetUint64(1), []byte{1})

	// add one more acc to world state
	worldState.Add(types.Address{2}, 1, new(big.Int).SetUint64(1), []byte{1})

	if worldState.Equal(comparedWorldStateEqual) {
		t.Fatal("world states are different but equal returned false")
	}
}

func TestWorld_Copy(t *testing.T) {
	hashOne := types.BigToHash(new(big.Int).SetUint64(1))
	hashTwo := types.BigToHash(new(big.Int).SetUint64(2))
	acc := NewAccount(1, new(big.Int).SetUint64(1), []byte{1})
	acc.Storage = make(map[types.Hash]types.Hash)
	acc.Storage[hashOne] = hashTwo

	cpy := acc.Copy()
	if !acc.Equal(cpy) {
		t.Fatalf("accounts values must be equal\ngot: %v\nwant: %v", cpy, acc)
	}
}

func TestWorldState_Diff(t *testing.T) {
	addr1 := types.Address{1}
	addr2 := types.Address{2}

	tests := []struct {
		name     string
		ws1      WorldState
		ws2      WorldState
		expected WorldState
	}{
		{
			name:     "empty states",
			ws1:      make(WorldState),
			ws2:      make(WorldState),
			expected: make(WorldState),
		},
		{
			name:     "account only in first state",
			ws1:      make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1}),
			ws2:      make(WorldState),
			expected: make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1}),
		},
		{
			name:     "same account different values",
			ws1:      make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1}),
			ws2:      make(WorldState).Add(addr1, 2, new(big.Int).SetUint64(2), []byte{2}),
			expected: make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1}),
		},
		{
			name:     "same account same values",
			ws1:      make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1}),
			ws2:      make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1}),
			expected: make(WorldState),
		},
		{
			name: "different storage values",
			ws1: func() WorldState {
				ws := make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
				ws[addr1].Storage[types.Hash{1}] = types.Hash{2}
				return ws
			}(),
			ws2: func() WorldState {
				ws := make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
				ws[addr1].Storage[types.Hash{1}] = types.Hash{3}
				return ws
			}(),
			expected: func() WorldState {
				ws := make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
				ws[addr1].Storage[types.Hash{1}] = types.Hash{2}
				return ws
			}(),
		},
		{
			name: "storage value exists only in first state",
			ws1: func() WorldState {
				ws := make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
				ws[addr1].Storage[types.Hash{1}] = types.Hash{2}
				return ws
			}(),
			ws2: make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1}),
			expected: func() WorldState {
				ws := make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
				ws[addr1].Storage[types.Hash{1}] = types.Hash{2}
				return ws
			}(),
		},
		{
			name: "multiple accounts with different storages",
			ws1: func() WorldState {
				ws := make(WorldState)
				ws.Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
				ws.Add(addr2, 2, new(big.Int).SetUint64(2), []byte{2})
				ws[addr1].Storage[types.Hash{1}] = types.Hash{2}
				ws[addr2].Storage[types.Hash{3}] = types.Hash{4}
				return ws
			}(),
			ws2: func() WorldState {
				ws := make(WorldState)
				ws.Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
				ws.Add(addr2, 2, new(big.Int).SetUint64(2), []byte{2})
				ws[addr1].Storage[types.Hash{1}] = types.Hash{3}
				return ws
			}(),
			expected: func() WorldState {
				ws := make(WorldState)
				ws.Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
				ws.Add(addr2, 2, new(big.Int).SetUint64(2), []byte{2})
				ws[addr1].Storage[types.Hash{1}] = types.Hash{2}
				ws[addr2].Storage[types.Hash{3}] = types.Hash{4}
				return ws
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ws1.Diff(tt.ws2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorldState_EstimateIncrementalSize(t *testing.T) {
	tests := []struct {
		name     string
		ws       WorldState
		y        WorldState
		expected uint64
	}{
		{
			name:     "empty world states",
			ws:       make(WorldState),
			y:        make(WorldState),
			expected: 0,
		},
		{
			name: "completely new account without storage",
			ws:   make(WorldState),
			y: func() WorldState {
				ws := make(WorldState)
				ws.Add(types.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
				return ws
			}(),
			// address + nonce + balance (1 byte) + codehash
			expected: sizeOfAddress + sizeOfNonce + 1 + sizeOfHash,
		},
		{
			name: "completely new account with storage",
			ws:   make(WorldState),
			y: func() WorldState {
				ws := make(WorldState)
				acc := ws.Add(types.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
				acc[types.Address{1}].Storage[types.Hash{1}] = types.Hash{2}
				acc[types.Address{1}].Storage[types.Hash{2}] = types.Hash{3}
				return ws
			}(),
			// (address + nonce + balance (1 byte) + codehash) + (2 storage slots * hash size)
			expected: (sizeOfAddress + sizeOfNonce + 1 + sizeOfHash) + (2 * sizeOfHash),
		},
		{
			name: "existing account with same values",
			ws: func() WorldState {
				return make(WorldState).Add(types.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
			}(),
			y: func() WorldState {
				return make(WorldState).Add(types.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
			}(),
			expected: 0,
		},
		{
			name: "existing account with new storage keys",
			ws: func() WorldState {
				ws := make(WorldState)
				acc := ws.Add(types.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
				acc[types.Address{1}].Storage[types.Hash{1}] = types.Hash{2}
				return ws
			}(),
			y: func() WorldState {
				ws := make(WorldState)
				acc := ws.Add(types.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
				acc[types.Address{1}].Storage[types.Hash{1}] = types.Hash{2}
				acc[types.Address{1}].Storage[types.Hash{2}] = types.Hash{3}
				return ws
			}(),
			// one new storage key
			expected: sizeOfHash,
		},
		{
			name: "multiple accounts with mixed scenarios",
			ws: func() WorldState {
				ws := make(WorldState)
				acc1 := ws.Add(types.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
				acc1[types.Address{1}].Storage[types.Hash{1}] = types.Hash{2}
				return ws
			}(),
			y: func() WorldState {
				ws := make(WorldState)
				// existing account with new storage
				acc1 := ws.Add(types.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
				acc1[types.Address{1}].Storage[types.Hash{1}] = types.Hash{2}
				acc1[types.Address{1}].Storage[types.Hash{2}] = types.Hash{3}
				// completely new account with storage
				acc2 := ws.Add(types.Address{2}, 2, new(big.Int).SetUint64(2), []byte{2})
				acc2[types.Address{2}].Storage[types.Hash{3}] = types.Hash{4}
				return ws
			}(),
			// (new storage key) + (address + nonce + balance (1 byte) + codehash + storage slot)
			expected: sizeOfHash + (sizeOfAddress + sizeOfNonce + 1 + sizeOfHash + sizeOfHash),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ws.EstimateIncrementalSize(tt.y)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestWorldState_Merge(t *testing.T) {
	tests := []struct {
		name     string
		ws       WorldState
		y        WorldState
		expected WorldState
	}{
		{
			name:     "empty states",
			ws:       make(WorldState),
			y:        make(WorldState),
			expected: make(WorldState),
		},
		{
			name: "merge new account",
			ws:   make(WorldState),
			y: func() WorldState {
				ws := make(WorldState)
				acc := ws.Add(types.Address{1}, 1, new(big.Int).SetUint64(100), []byte{1})
				acc[types.Address{1}].Storage[types.Hash{1}] = types.Hash{2}
				return ws
			}(),
			expected: func() WorldState {
				ws := make(WorldState)
				acc := ws.Add(types.Address{1}, 1, new(big.Int).SetUint64(100), []byte{1})
				acc[types.Address{1}].Storage[types.Hash{1}] = types.Hash{2}
				return ws
			}(),
		},
		{
			name: "merge existing account with same values",
			ws: func() WorldState {
				return make(WorldState).Add(types.Address{1}, 1, new(big.Int).SetUint64(100), []byte{1})
			}(),
			y: func() WorldState {
				return make(WorldState).Add(types.Address{1}, 1, new(big.Int).SetUint64(100), []byte{1})
			}(),
			expected: func() WorldState {
				return make(WorldState).Add(types.Address{1}, 1, new(big.Int).SetUint64(100), []byte{1})
			}(),
		},
		{
			name: "merge existing account with different values",
			ws: func() WorldState {
				return make(WorldState).Add(types.Address{1}, 1, new(big.Int).SetUint64(100), []byte{1})
			}(),
			y: func() WorldState {
				return make(WorldState).Add(types.Address{1}, 2, new(big.Int).SetUint64(200), []byte{2})
			}(),
			expected: func() WorldState {
				return make(WorldState).Add(types.Address{1}, 2, new(big.Int).SetUint64(200), []byte{2})
			}(),
		},
		{
			name: "merge account with storage",
			ws: func() WorldState {
				ws := make(WorldState).Add(types.Address{1}, 1, new(big.Int).SetUint64(100), []byte{1})
				ws[types.Address{1}].Storage[types.Hash{1}] = types.Hash{1}
				return ws
			}(),
			y: func() WorldState {
				ws := make(WorldState).Add(types.Address{1}, 1, new(big.Int).SetUint64(100), []byte{1})
				ws[types.Address{1}].Storage[types.Hash{1}] = types.Hash{2}
				ws[types.Address{1}].Storage[types.Hash{2}] = types.Hash{3}
				return ws
			}(),
			expected: func() WorldState {
				ws := make(WorldState).Add(types.Address{1}, 1, new(big.Int).SetUint64(100), []byte{1})
				ws[types.Address{1}].Storage[types.Hash{1}] = types.Hash{2}
				ws[types.Address{1}].Storage[types.Hash{2}] = types.Hash{3}
				return ws
			}(),
		},
		{
			name: "merge multiple accounts with storage",
			ws: func() WorldState {
				ws := make(WorldState)
				acc1 := ws.Add(types.Address{1}, 1, new(big.Int).SetUint64(100), []byte{1})
				acc1[types.Address{1}].Storage[types.Hash{1}] = types.Hash{1}
				return ws
			}(),
			y: func() WorldState {
				ws := make(WorldState)
				acc1 := ws.Add(types.Address{1}, 2, new(big.Int).SetUint64(200), []byte{2})
				acc1[types.Address{1}].Storage[types.Hash{1}] = types.Hash{2}
				acc2 := ws.Add(types.Address{2}, 3, new(big.Int).SetUint64(300), []byte{3})
				acc2[types.Address{2}].Storage[types.Hash{3}] = types.Hash{4}
				return ws
			}(),
			expected: func() WorldState {
				ws := make(WorldState)
				acc1 := ws.Add(types.Address{1}, 2, new(big.Int).SetUint64(200), []byte{2})
				acc1[types.Address{1}].Storage[types.Hash{1}] = types.Hash{2}
				acc2 := ws.Add(types.Address{2}, 3, new(big.Int).SetUint64(300), []byte{3})
				acc2[types.Address{2}].Storage[types.Hash{3}] = types.Hash{4}
				return ws
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ws.Merge(tt.y)
			assert.Equal(t, tt.ws, tt.expected)
		})
	}
}
