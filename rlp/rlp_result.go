package rlp

import (
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
)

func NewResult(result *substate.Result) *Result {
	return &Result{
		Status:          result.Status,
		Bloom:           result.Bloom,
		Logs:            result.Logs,
		ContractAddress: result.ContractAddress,
		GasUsed:         result.GasUsed,
	}
}

type Result struct {
	Status uint64
	Bloom  types.Bloom
	Logs   []*types.Log

	ContractAddress types.Address
	GasUsed         uint64
}

// ToSubstate transforms r from Result to substate.Result.
func (r Result) ToSubstate() *substate.Result {
	return &substate.Result{
		Status:          r.Status,
		Bloom:           r.Bloom,
		Logs:            r.Logs,
		ContractAddress: r.ContractAddress,
		GasUsed:         r.GasUsed,
	}
}
