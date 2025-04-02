package types

import (
	"math/big"
)

// SetCodeAuthorization authorization from an account to deploy code at its address.
type SetCodeAuthorization struct {
	ChainID big.Int
	Address Address
	Nonce   uint64
	V       uint8
	R       big.Int
	S       big.Int
}
