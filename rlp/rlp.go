package rlp

import (
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/types/rlp"
)

func NewRLP(substate *substate.Substate) (*RLP, error) {
	input, err := NewWorldState(substate.InputSubstate)
	if err != nil {
		return nil, err
	}
	output, err := NewWorldState(substate.OutputSubstate)
	if err != nil {
		return nil, err
	}
	message, err := NewMessage(substate.Message)
	if err != nil {
		return nil, err
	}
	return &RLP{
		InputSubstate:  input,
		OutputSubstate: output,
		Env:            NewEnv(substate.Env),
		Message:        message,
		Result:         NewResult(substate.Result),
	}, nil
}

type RLP struct {
	InputSubstate  WorldState
	OutputSubstate WorldState
	Env            *Env
	Message        *Message
	Result         *Result
}

// Decode decodes val into RLP and returns it.
func Decode(val []byte) (*RLP, error) {
	var err error

	// londonRLP has currently the biggest representation the DB, so it should always be first.
	var london londonRLP
	err = rlp.DecodeBytes(val, &london)
	if err == nil {
		return london.toRLP(), nil
	}

	var berlin berlinRLP
	err = rlp.DecodeBytes(val, &berlin)
	if err == nil {
		return berlin.toRLP(), nil
	}

	var legacy legacySubstateRLP
	err = rlp.DecodeBytes(val, &legacy)
	if err == nil {
		return legacy.toRLP(), nil
	}

	// cancun
	var substateRLP RLP
	err = rlp.DecodeBytes(val, &substateRLP)
	if err == nil {
		return &substateRLP, nil
	}

	return nil, err
}

// ToSubstate transforms every attribute of r from RLP to substate.Substate.
func (r *RLP) ToSubstate(getHashFunc func(codeHash types.Hash) ([]byte, error), block uint64, tx int) (*substate.Substate, error) {
	msg, err := r.Message.ToSubstate(getHashFunc)
	if err != nil {
		return nil, err
	}

	input, err := r.InputSubstate.ToSubstate(getHashFunc)
	if err != nil {
		return nil, err
	}
	output, err := r.OutputSubstate.ToSubstate(getHashFunc)
	if err != nil {
		return nil, err
	}
	return &substate.Substate{
		InputSubstate:  input,
		OutputSubstate: output,
		Env:            r.Env.ToSubstate(),
		Message:        msg,
		Result:         r.Result.ToSubstate(),
		Block:          block,
		Transaction:    tx,
	}, nil
}
