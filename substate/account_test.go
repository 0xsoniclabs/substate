package substate

import (
	"math/big"
	"testing"

	"github.com/holiman/uint256"

	"github.com/0xsoniclabs/substate/types/hash"
	"github.com/stretchr/testify/assert"

	"github.com/0xsoniclabs/substate/types"
)

func TestAccount_EqualNonce(t *testing.T) {
	acc := NewAccount(1, new(uint256.Int).SetUint64(1), []byte{1})
	comparedNonceAcc := NewAccount(2, new(uint256.Int).SetUint64(1), []byte{1})

	if acc.Equal(comparedNonceAcc) {
		t.Fatal("accounts nonce are different but equal returned true")
	}

	comparedNonceAcc.Nonce = acc.Nonce
	if !acc.Equal(comparedNonceAcc) {
		t.Fatal("accounts nonce are same but equal returned false")
	}
}

func TestAccount_EqualBalance(t *testing.T) {
	acc := NewAccount(1, new(uint256.Int).SetUint64(1), []byte{1})
	comparedBalanceAcc := NewAccount(1, new(uint256.Int).SetUint64(2), []byte{1})

	if acc.Equal(comparedBalanceAcc) {
		t.Fatal("accounts balances are different but equal returned true")
	}

	comparedBalanceAcc.Balance.SetUint64(acc.Balance.Uint64())
	if !acc.Equal(comparedBalanceAcc) {
		t.Fatal("accounts balances are same but equal returned false")
	}
}

func TestAccount_EqualStorage(t *testing.T) {
	hashOne := types.BigToHash(new(big.Int).SetUint64(1))
	hashTwo := types.BigToHash(new(big.Int).SetUint64(2))
	hashThree := types.BigToHash(new(big.Int).SetUint64(3))

	acc := NewAccount(1, new(uint256.Int).SetUint64(1), []byte{1})
	acc.Storage = make(map[types.Hash]types.Hash)
	acc.Storage[hashOne] = hashTwo

	// first compare with no storage
	comparedStorageAcc := NewAccount(1, new(uint256.Int).SetUint64(1), []byte{1})
	if acc.Equal(comparedStorageAcc) {
		t.Fatal("accounts storages are different but equal returned true")
	}

	// then compare different value for same key
	comparedStorageAcc.Storage = make(map[types.Hash]types.Hash)
	comparedStorageAcc.Storage[hashOne] = hashThree
	if acc.Equal(comparedStorageAcc) {
		t.Fatal("accounts storages are different but equal returned true")
	}

	// then compare different keys
	comparedStorageAcc.Storage = make(map[types.Hash]types.Hash)
	comparedStorageAcc.Storage[hashTwo] = hashThree
	if acc.Equal(comparedStorageAcc) {
		t.Fatal("accounts storages are different but equal returned true")
	}

	// then compare same
	comparedStorageAcc.Storage = make(map[types.Hash]types.Hash)
	comparedStorageAcc.Storage[hashOne] = hashTwo

	if !acc.Equal(comparedStorageAcc) {
		t.Fatal("accounts storages are same but equal returned false")
	}
}

func TestAccount_EqualCode(t *testing.T) {
	acc := NewAccount(1, new(uint256.Int).SetUint64(1), []byte{1})
	comparedCodeAcc := NewAccount(1, new(uint256.Int).SetUint64(1), []byte{2})
	if acc.Equal(comparedCodeAcc) {
		t.Fatal("accounts codes are different but equal returned true")
	}

	copy(comparedCodeAcc.Code, acc.Code)
	if !acc.Equal(comparedCodeAcc) {
		t.Fatal("accounts codes are same but equal returned false")
	}

}

func TestAccount_Copy(t *testing.T) {
	hashOne := types.BigToHash(new(big.Int).SetUint64(1))
	hashTwo := types.BigToHash(new(big.Int).SetUint64(2))
	acc := NewAccount(1, new(uint256.Int).SetUint64(1), []byte{1})
	acc.Storage = make(map[types.Hash]types.Hash)
	acc.Storage[hashOne] = hashTwo

	cpy := acc.Copy()
	if !acc.Equal(cpy) {
		t.Fatalf("accounts values must be equal\ngot: %v\nwant: %v", cpy, acc)
	}
}

func TestAccount_CodeHash(t *testing.T) {
	acc := NewAccount(1, new(uint256.Int).SetUint64(1), []byte{1})
	expectedHash := hash.Keccak256Hash(acc.Code)
	assert.Equal(t, expectedHash, acc.CodeHash())
}

func TestAccount_String(t *testing.T) {
	hashOne := types.BigToHash(new(big.Int).SetUint64(1))
	hashTwo := types.BigToHash(new(big.Int).SetUint64(2))
	acc := NewAccount(1, new(uint256.Int).SetUint64(1), []byte{1})
	acc.Storage = make(map[types.Hash]types.Hash)
	acc.Storage[hashOne] = hashTwo
	expectedString := "Nonce: 1\nBalance: 1\nCode: \x01\nStorage:0x0000000000000000000000000000000000000000000000000000000000000001: 0x0000000000000000000000000000000000000000000000000000000000000002\n"
	assert.Equal(t, expectedString, acc.String())
}

func TestAccount_EqualSelf(t *testing.T) {
	acc := NewAccount(1, new(uint256.Int).SetUint64(1), []byte{1})
	assert.Equal(t, true, acc.Equal(acc))
}

func TestAccount_EqualNil(t *testing.T) {
	acc := NewAccount(1, new(uint256.Int).SetUint64(1), []byte{1})
	assert.Equal(t, false, acc.Equal(nil))
}
