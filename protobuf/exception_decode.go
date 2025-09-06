package protobuf

import (
	"github.com/0xsoniclabs/substate/substate"
)

// Decode decodes protobuf-encoded bytes into ExceptionBlock struct
func (s *ExceptionBlock) Decode(lookup getCodeFunc) (*substate.ExceptionBlock, error) {
	input, err := s.GetPreBlock().decode(lookup)
	if err != nil {
		return nil, err
	}

	output, err := s.GetPostBlock().decode(lookup)
	if err != nil {
		return nil, err
	}

	dataRaw := s.GetTransactions()
	txs := make(map[int]substate.ExceptionTx, len(dataRaw))
	for key, value := range dataRaw {
		txs[int(key)] = value.decode(lookup)
	}

	return &substate.ExceptionBlock{
		Transactions: txs,
		PreBlock:     input,
		PostBlock:    output,
	}, nil
}

// decode decodes protobuf-encoded ExceptionTx into ExceptionTx struct
func (tx *ExceptionTx) decode(lookup getCodeFunc) substate.ExceptionTx {
	preTransaction, err := tx.GetPreTransaction().decode(lookup)
	if err != nil {
		return substate.ExceptionTx{}
	}

	postTransaction, err := tx.GetPostTransaction().decode(lookup)
	if err != nil {
		return substate.ExceptionTx{}
	}

	return substate.ExceptionTx{
		PreTransaction:  preTransaction,
		PostTransaction: postTransaction,
		VmException:     tx.GetVmException(),
	}
}
