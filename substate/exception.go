package substate

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

func (ex *Exception) Equal(target Exception) bool {
	// Check if the block numbers are the same
	if ex.Block != target.Block {
		return false
	}

	// Check if the contecnt of exceptions are the same
	if !ex.Data.Equal(target.Data) {
		return false
	}

	return true
}

// Equal checks if two ExceptionBlock instances are equal
func (ex *ExceptionBlock) Equal(target ExceptionBlock) bool {
	if len(ex.Transactions) != len(target.Transactions) {
		return false
	}

	for i, tx := range ex.Transactions {
		if !tx.Equal(target.Transactions[i]) {
			return false
		}
	}

	if !ex.PreBlock.Equal(*target.PreBlock) {
		return false
	}

	if !ex.PostBlock.Equal(*target.PostBlock) {
		return false
	}

	return true
}

// Equal checks if two ExceptionTx instances are equal
func (tx *ExceptionTx) Equal(target ExceptionTx) bool {
	if !tx.PreTransaction.Equal(*target.PreTransaction) {
		return false
	}

	if !tx.PostTransaction.Equal(*target.PostTransaction) {
		return false
	}

	if tx.VmException != target.VmException {
		return false
	}

	return true
}
