// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package hash

import (
	"fmt"
	"hash"

	"golang.org/x/crypto/sha3"

	"github.com/0xsoniclabs/substate/types"
)

// KeccakState wraps sha3.state. In addition to the usual hash methods, it also supports
// Read to get a variable amount of data from the hash state. Read is faster than Sum
// because it doesn't copy the internal state, but also modifies the internal state.
//
//go:generate mockgen -source=hash.go -destination=./hash_mock.go -package=hash
type KeccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}

// NewKeccakState creates a new KeccakState
func NewKeccakState() KeccakState {
	return sha3.NewLegacyKeccak256().(KeccakState)
}

// Keccak256Hash calculates and returns the Keccak256 hash of the input data,
// converting it to an internal Hash data structure.
func Keccak256Hash(data ...[]byte) (h types.Hash) {
	d := NewKeccakState()
	for _, b := range data {
		_, err := d.Write(b)
		if err != nil {
			panic(fmt.Errorf("failed to write to keccak256 hash: %v", err))
		}
	}
	_, err := d.Read(h[:])
	if err != nil {
		panic(fmt.Errorf("failed to read from keccak256 hash: %v", err))
	}
	return h
}
