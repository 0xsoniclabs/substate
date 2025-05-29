package protobuf

import (
	"github.com/0xsoniclabs/substate/substate"
	"google.golang.org/protobuf/proto"
)

// EncodeExceptionBlock converts aida ExceptionBlock into protobuf-encoded ExceptionBlock
func EncodeExceptionBlock(s *substate.ExceptionBlock) ([]byte, error) {
	data := make(map[int32]*ExceptionTx, len(s.Transactions))
	for key, value := range s.Transactions {
		encodedTx, err := EncodeExceptionTx(&value)
		if err != nil {
			return nil, err
		}
		data[int32(key)] = encodedTx
	}

	var pre *Alloc
	if s.PreBlock != nil {
		pre = toProtobufAlloc(*s.PreBlock)
	}

	var post *Alloc
	if s.PostBlock != nil {
		post = toProtobufAlloc(*s.PostBlock)
	}

	return proto.Marshal(&ExceptionBlock{
		Transactions: data,
		PreBlock:     pre,
		PostBlock:    post,
	})
}

// EncodeExceptionTx converts aida ExceptionTx into protobuf-encoded ExceptionTx
func EncodeExceptionTx(tx *substate.ExceptionTx) (*ExceptionTx, error) {
	var pre *Alloc
	if tx.PreTransaction != nil {
		pre = toProtobufAlloc(*tx.PreTransaction)
	}
	var post *Alloc
	if tx.PostTransaction != nil {
		post = toProtobufAlloc(*tx.PostTransaction)
	}
	return &ExceptionTx{
		PreTransaction:  pre,
		PostTransaction: post,
		VmException:     &tx.VmException,
	}, nil
}
