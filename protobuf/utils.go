package protobuf

import (
	"math/big"

	"github.com/0xsoniclabs/substate/types/hash"
	"google.golang.org/protobuf/proto"

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

func BigIntToBytes(i *big.Int) []byte {
	if i == nil {
		return nil
	}
	return i.Bytes()
}

// CodeHash computes the Keccak256 hash of the given byte slice `code`.
//
// Parameters:
// - code: A byte slice representing the code to be hashed.
//
// Returns:
// - A types.Hash object containing the Keccak256 hash of the input code.
func CodeHash(code []byte) types.Hash {
	return hash.Keccak256Hash(code)
}

// HashToBytes converts a types.Hash object to a byte slice.
//
// Parameters:
// - hash: A pointer to a types.Hash object to be converted.
//
// Returns:
// - A byte slice representing the hash. If the input hash is nil, it returns nil.
func HashToBytes(hash *types.Hash) []byte {
	if hash == nil {
		return nil
	}
	return hash.Bytes()
}

// HashedCopy creates a deep copy of the Substate object with code hashes instead of code bytes.
//
// Returns:
// - A new Substate object with code hashes instead of code bytes.
func (s *Substate) HashedCopy() *Substate {
	y := proto.Clone(s).(*Substate)

	if y == nil {
		return nil
	}

	for _, entry := range y.InputAlloc.Alloc {
		account := entry.Account
		if code := account.GetCode(); code != nil {
			codeHash := CodeHash(code)
			account.Contract = &Substate_Account_CodeHash{
				CodeHash: HashToBytes(&codeHash),
			}
		}
	}

	for _, entry := range y.OutputAlloc.Alloc {
		account := entry.Account
		if code := account.GetCode(); code != nil {
			codeHash := CodeHash(code)
			account.Contract = &Substate_Account_CodeHash{
				CodeHash: HashToBytes(&codeHash),
			}
		}
	}

	if y.TxMessage.To == nil {
		if code := y.TxMessage.GetData(); code != nil {
			codeHash := CodeHash(code)
			y.TxMessage.Input = &Substate_TxMessage_InitCodeHash{
				InitCodeHash: HashToBytes(&codeHash),
			}
		}
	}

	return y
}
