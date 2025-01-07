package protobuf

import (
	"math/big"

	"github.com/0xsoniclabs/substate/types"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

func BytesValueToHash(bv *wrapperspb.BytesValue) *types.Hash {
	if bv == nil {
		return nil
	}
	hash := types.BytesToHash(bv.GetValue())
	return &hash
}

func BytesValueToBigInt(bv *wrapperspb.BytesValue) *big.Int {
	if bv == nil {
		return nil
	}
	return new(big.Int).SetBytes(bv.GetValue())
}

func BytesValueToAddress(bv *wrapperspb.BytesValue) *types.Address {
	if bv == nil {
		return nil
	}
	addr := types.BytesToAddress(bv.GetValue())
	return &addr
}

func AddressToWrapperspbBytes(a *types.Address) *wrapperspb.BytesValue {
	if a == nil {
		return nil
	}
	return wrapperspb.Bytes(a.Bytes())
}

func HashToWrapperspbBytes(h *types.Hash) *wrapperspb.BytesValue {
	if h == nil {
		return nil
	}
	return wrapperspb.Bytes(h.Bytes())
}

func BigIntToWrapperspbBytes(i *big.Int) *wrapperspb.BytesValue {
	if i == nil {
		return nil
	}
	return wrapperspb.Bytes(i.Bytes())
}

func BytesToBigInt(b []byte) *big.Int {
	if b == nil {
		return nil
	}
	return new(big.Int).SetBytes(b)
}
