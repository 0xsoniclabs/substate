package substate

import (
	"testing"

	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
)

type exceptionBlockEqualTestCase struct {
	name        string
	eb1         ExceptionBlock
	eb2         ExceptionBlock
	expectedErr string
}

func TestExceptionBlock_Equal_Cases(t *testing.T) {
	ws := &WorldState{}
	ws1 := &WorldState{}
	ws2 := &WorldState{types.Address{1}: &Account{Nonce: 1}}
	tx1 := ExceptionTx{PreTransaction: ws, PostTransaction: ws, VmException: false}
	tx2 := ExceptionTx{PreTransaction: ws, PostTransaction: ws, VmException: true}

	cases := []exceptionBlockEqualTestCase{
		{
			name:        "TransactionCountMismatch",
			eb1:         ExceptionBlock{Transactions: map[int]ExceptionTx{0: {PreTransaction: ws, PostTransaction: ws}}, PreBlock: ws, PostBlock: ws},
			eb2:         ExceptionBlock{Transactions: map[int]ExceptionTx{}, PreBlock: ws, PostBlock: ws},
			expectedErr: "transaction count mismatch",
		},
		{
			name:        "TransactionMismatch",
			eb1:         ExceptionBlock{Transactions: map[int]ExceptionTx{0: tx1}, PreBlock: ws, PostBlock: ws},
			eb2:         ExceptionBlock{Transactions: map[int]ExceptionTx{0: tx2}, PreBlock: ws, PostBlock: ws},
			expectedErr: "VM exception mismatch",
		},
		{
			name:        "PreBlockMismatch",
			eb1:         ExceptionBlock{Transactions: map[int]ExceptionTx{}, PreBlock: ws1, PostBlock: ws1},
			eb2:         ExceptionBlock{Transactions: map[int]ExceptionTx{}, PreBlock: ws2, PostBlock: ws1},
			expectedErr: "pre block mismatch",
		},
		{
			name:        "PostBlockMismatch",
			eb1:         ExceptionBlock{Transactions: map[int]ExceptionTx{}, PreBlock: ws1, PostBlock: ws1},
			eb2:         ExceptionBlock{Transactions: map[int]ExceptionTx{}, PreBlock: ws1, PostBlock: ws2},
			expectedErr: "post block mismatch",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.eb1.Equal(tc.eb2)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

type exceptionTxEqualTestCase struct {
	name        string
	tx1         ExceptionTx
	tx2         ExceptionTx
	expectedErr string
}

func TestExceptionTx_Equal_Cases(t *testing.T) {
	ws := &WorldState{}
	ws1 := &WorldState{}
	ws2 := &WorldState{types.Address{1}: &Account{Nonce: 1}}

	cases := []exceptionTxEqualTestCase{
		{
			name:        "PreTransactionMismatch",
			tx1:         ExceptionTx{PreTransaction: ws1, PostTransaction: ws1, VmException: false},
			tx2:         ExceptionTx{PreTransaction: ws2, PostTransaction: ws1, VmException: false},
			expectedErr: "pre transaction mismatch",
		},
		{
			name:        "PostTransactionMismatch",
			tx1:         ExceptionTx{PreTransaction: ws1, PostTransaction: ws1, VmException: false},
			tx2:         ExceptionTx{PreTransaction: ws1, PostTransaction: ws2, VmException: false},
			expectedErr: "post transaction mismatch",
		},
		{
			name:        "VmExceptionMismatch",
			tx1:         ExceptionTx{PreTransaction: ws, PostTransaction: ws, VmException: false},
			tx2:         ExceptionTx{PreTransaction: ws, PostTransaction: ws, VmException: true},
			expectedErr: "VM exception mismatch",
		},
		{
			name:        "Success",
			tx1:         ExceptionTx{PreTransaction: ws, PostTransaction: ws, VmException: false},
			tx2:         ExceptionTx{PreTransaction: ws, PostTransaction: ws, VmException: false},
			expectedErr: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.tx1.Equal(tc.tx2)
			if tc.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}
