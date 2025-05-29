package substate

import "fmt"

type Exception struct {
	Block uint64
	Data  ExceptionBlock
}

type ExceptionBlock struct {
	Transactions map[int]ExceptionTx
	PreBlock     *WorldState
	PostBlock    *WorldState
}

type ExceptionTx struct {
	PreTransaction  *WorldState
	PostTransaction *WorldState
	VmException     bool
}

func (ex *Exception) Equal(target Exception) error {
	if ex.Block != target.Block {
		return fmt.Errorf("block mismatch: got %d, want %d", ex.Block, target.Block)
	}

	if err := ex.Data.Equal(target.Data); err != nil {
		return fmt.Errorf("exception block mismatch: %w", err)
	}

	return nil
}

func (ex *ExceptionBlock) Equal(target ExceptionBlock) error {
	if len(ex.Transactions) != len(target.Transactions) {
		return fmt.Errorf("transaction count mismatch: got %d, want %d", len(ex.Transactions), len(target.Transactions))
	}

	for i, t := range ex.Transactions {
		if err := t.Equal(target.Transactions[i]); err != nil {
			return fmt.Errorf("transaction %d mismatch: %w", i, err)
		}
	}

	if !ex.PreBlock.Equal(*target.PreBlock) {
		return fmt.Errorf("pre block mismatch")
	}

	if !ex.PostBlock.Equal(*target.PostBlock) {
		return fmt.Errorf("post block mismatch")
	}

	return nil
}

func (tx *ExceptionTx) Equal(target ExceptionTx) error {
	if !tx.PreTransaction.Equal(*target.PreTransaction) {
		return fmt.Errorf("pre transaction mismatch")
	}

	if !tx.PostTransaction.Equal(*target.PostTransaction) {
		return fmt.Errorf("post transaction mismatch")
	}

	if tx.VmException != target.VmException {
		return fmt.Errorf("VM exception mismatch: got %v, want %v", tx.VmException, target.VmException)
	}

	return nil
}
