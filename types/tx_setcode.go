package types

import (
	"github.com/holiman/uint256"
)

// SetCodeAuthorization authorization from an account to deploy code at its address.
type SetCodeAuthorization struct {
	ChainID *uint256.Int
	Address Address
	Nonce   uint64
	V       uint8        // signature parity
	R       *uint256.Int // signature R parameter
	S       *uint256.Int // signature S parameter
}

// Equal checks if two SetCodeAuthorization are equal.
func (s *SetCodeAuthorization) Equal(to SetCodeAuthorization) bool {
	// check addresses position
	if s.Address != to.Address {
		return false
	}

	// check chainId position
	if s.ChainID.Cmp(to.ChainID) != 0 {
		return false
	}

	// check nonce position
	if s.Nonce != to.Nonce {
		return false
	}

	if s.V != to.V {
		return false
	}

	if s.R.Cmp(to.R) != 0 {
		return false
	}

	if s.S.Cmp(to.S) != 0 {
		return false
	}
	return true
}
