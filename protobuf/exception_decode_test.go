package protobuf

import (
	"math/big"
	"testing"

	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/runtime/protoimpl"
)

func TestExceptionDecode(t *testing.T) {
	lookup := func(hash types.Hash) ([]byte, error) {
		return nil, nil
	}

	alloc := []*AllocEntry{
		{
			Address: []byte{2},
			Account: &Account{
				state:   protoimpl.MessageState{},
				Nonce:   uint64Ptr(2),
				Balance: big.NewInt(200).Bytes(),
				Storage: []*Account_StorageEntry{
					{
						Key:   []byte{1},
						Value: []byte{1},
					},
				},
			},
		},
	}

	vmException := true

	// Create a sample ExceptionBlock with minimal valid data
	exBlock := &ExceptionBlock{
		Transactions: map[int32]*ExceptionTx{
			1: {
				PreTransaction:  &Alloc{Alloc: alloc},
				PostTransaction: &Alloc{Alloc: alloc},
				VmException:     &vmException,
			},
		},
		PreBlock:  &Alloc{Alloc: alloc},
		PostBlock: &Alloc{Alloc: alloc},
	}

	d, err := exBlock.Decode(lookup)
	assert.NoError(t, err)

	expected := d.PreBlock.String()
	if d.PreBlock.String() != expected {
		t.Errorf("Expected PreBlock to be %s, got %s", expected, d.PreBlock.String())
	}
	if d.PostBlock.String() != expected {
		t.Errorf("Expected PostBlock to be %s, got %s", expected, d.PostBlock.String())
	}
	if len(d.Transactions) != 1 {
		t.Errorf("Expected Data to have 1 entry, got %d", len(d.Transactions))
	}
	assert.Equal(t, 1, len(d.Transactions))
	for _, tx := range d.Transactions {
		if tx.PreTransaction.String() != expected {
			t.Errorf("Expected PreTransaction to be %s, got %s", expected, tx.PreTransaction.String())
		}
		if tx.PostTransaction.String() != expected {
			t.Errorf("Expected PostTransaction to be %s, got %s", expected, tx.PostTransaction.String())
		}
		assert.Equal(t, true, tx.VmException)
	}
	assert.NoError(t, err)
}
