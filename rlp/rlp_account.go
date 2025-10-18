package rlp

import (
	"sort"

	"github.com/0xsoniclabs/substate/utils"
	"github.com/holiman/uint256"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
)

func NewRLPAccount(acc *substate.Account) *SubstateAccountRLP {
	a := &SubstateAccountRLP{
		Nonce:    acc.Nonce,
		Balance:  new(uint256.Int).Set(acc.Balance),
		CodeHash: utils.Must(acc.CodeHash()),
		Storage:  [][2]types.Hash{},
	}

	var sortedKeys []types.Hash
	for key := range acc.Storage {
		sortedKeys = append(sortedKeys, key)
	}

	sort.Slice(sortedKeys, func(i, j int) bool {
		return sortedKeys[i].Compare(sortedKeys[j]) < 0
	})

	for _, key := range sortedKeys {
		value := acc.Storage[key]
		a.Storage = append(a.Storage, [2]types.Hash{key, value})
	}

	return a
}

type SubstateAccountRLP struct {
	Nonce    uint64
	Balance  *uint256.Int
	CodeHash types.Hash
	Storage  [][2]types.Hash
}
