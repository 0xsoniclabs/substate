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
