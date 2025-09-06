package rlp

import (
	"errors"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/syndtr/goleveldb/leveldb"
)

type UpdateSetRLP struct {
	WorldState      WorldState
	DeletedAccounts []types.Address
}

func NewUpdateSetRLP(ws substate.WorldState, deletedAccounts []types.Address) *UpdateSetRLP {
	a := WorldState{
		Addresses: []types.Address{},
		Accounts:  []*SubstateAccountRLP{},
	}
	for addr, acc := range ws {
		a.Addresses = append(a.Addresses, addr)
		a.Accounts = append(a.Accounts, NewRLPAccount(acc))
	}
	return &UpdateSetRLP{
		WorldState:      a,
		DeletedAccounts: deletedAccounts,
	}
}

func (up *UpdateSetRLP) ToWorldState(lookup func(codeHash types.Hash) ([]byte, error)) (*substate.WorldState, error) {
	worldState := make(substate.WorldState)

	for i, addr := range up.WorldState.Addresses {
		worldStateAcc := up.WorldState.Accounts[i]

		code, err := lookup(worldStateAcc.CodeHash)
		if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
			return nil, err
		}

		acc := substate.Account{
			Nonce:   worldStateAcc.Nonce,
			Balance: worldStateAcc.Balance,
			Storage: make(map[types.Hash]types.Hash),
			Code:    code,
		}

		for j := range worldStateAcc.Storage {
			acc.Storage[worldStateAcc.Storage[j][0]] = worldStateAcc.Storage[j][1]
		}
		worldState[addr] = &acc
	}
	return &worldState, nil
}
