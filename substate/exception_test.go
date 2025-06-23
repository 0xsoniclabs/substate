package substate

import (
	"testing"

	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
)

type exceptionBlockEqualTestCase struct {
	name           string
	eb1            ExceptionBlock
	eb2            ExceptionBlock
	expectedResult bool
}

func TestExceptionBlock_Equal_Cases(t *testing.T) {
	ws := &WorldState{}
	ws1 := &WorldState{}
	ws2 := &WorldState{types.Address{1}: &Account{Nonce: 1}}
	tx1 := ExceptionTx{PreTransaction: ws, PostTransaction: ws, VmException: false}
	tx2 := ExceptionTx{PreTransaction: ws, PostTransaction: ws, VmException: true}

	cases := []exceptionBlockEqualTestCase{
		{
			name:           "TransactionCountMismatch",
			eb1:            ExceptionBlock{Transactions: map[int]ExceptionTx{0: {PreTransaction: ws, PostTransaction: ws}}, PreBlock: ws, PostBlock: ws},
			eb2:            ExceptionBlock{Transactions: map[int]ExceptionTx{}, PreBlock: ws, PostBlock: ws},
			expectedResult: false,
		},
		{
			name:           "TransactionMismatch",
			eb1:            ExceptionBlock{Transactions: map[int]ExceptionTx{0: tx1}, PreBlock: ws, PostBlock: ws},
			eb2:            ExceptionBlock{Transactions: map[int]ExceptionTx{0: tx2}, PreBlock: ws, PostBlock: ws},
			expectedResult: false,
		},
		{
			name:           "PreBlockMismatch",
			eb1:            ExceptionBlock{Transactions: map[int]ExceptionTx{}, PreBlock: ws1, PostBlock: ws1},
			eb2:            ExceptionBlock{Transactions: map[int]ExceptionTx{}, PreBlock: ws2, PostBlock: ws1},
			expectedResult: false,
		},
		{
			name:           "PostBlockMismatch",
			eb1:            ExceptionBlock{Transactions: map[int]ExceptionTx{}, PreBlock: ws1, PostBlock: ws1},
			eb2:            ExceptionBlock{Transactions: map[int]ExceptionTx{}, PreBlock: ws1, PostBlock: ws2},
			expectedResult: false,
		},
		{
			name:           "BlockMatch",
			eb1:            ExceptionBlock{Transactions: map[int]ExceptionTx{}, PreBlock: ws1, PostBlock: ws1},
			eb2:            ExceptionBlock{Transactions: map[int]ExceptionTx{}, PreBlock: ws1, PostBlock: ws1},
			expectedResult: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ret := tc.eb1.Equal(tc.eb2)
			assert.Equal(t, tc.expectedResult, ret)
		})
	}
}

type exceptionTxEqualTestCase struct {
	name           string
	tx1            ExceptionTx
	tx2            ExceptionTx
	expectedResult bool
}

func TestExceptionTx_Equal_Cases(t *testing.T) {
	ws := &WorldState{}
	ws1 := &WorldState{}
	ws2 := &WorldState{types.Address{1}: &Account{Nonce: 1}}

	cases := []exceptionTxEqualTestCase{
		{
			name:           "PreTransactionMismatch",
			tx1:            ExceptionTx{PreTransaction: ws1, PostTransaction: ws1, VmException: false},
			tx2:            ExceptionTx{PreTransaction: ws2, PostTransaction: ws1, VmException: false},
			expectedResult: false,
		},
		{
			name:           "PostTransactionMismatch",
			tx1:            ExceptionTx{PreTransaction: ws1, PostTransaction: ws1, VmException: false},
			tx2:            ExceptionTx{PreTransaction: ws1, PostTransaction: ws2, VmException: false},
			expectedResult: false,
		},
		{
			name:           "VmExceptionMismatch",
			tx1:            ExceptionTx{PreTransaction: ws, PostTransaction: ws, VmException: false},
			tx2:            ExceptionTx{PreTransaction: ws, PostTransaction: ws, VmException: true},
			expectedResult: false,
		},
		{
			name:           "Success",
			tx1:            ExceptionTx{PreTransaction: ws, PostTransaction: ws, VmException: false},
			tx2:            ExceptionTx{PreTransaction: ws, PostTransaction: ws, VmException: false},
			expectedResult: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ret := tc.tx1.Equal(tc.tx2)
			assert.Equal(t, tc.expectedResult, ret)
		})
	}
}
