package protobuf

import (
	"github.com/0xsoniclabs/substate/substate"
	"google.golang.org/protobuf/proto"
)

// EncodeExceptionBlock converts ExceptionBlock struct into protobuf-encoded ExceptionBlock
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
	var err error
	if s.PreBlock != nil {
		pre, err = toProtobufAlloc(*s.PreBlock)
		if err != nil {
			return nil, err
		}
	}

	var post *Alloc
	if s.PostBlock != nil {
		post, err = toProtobufAlloc(*s.PostBlock)
		if err != nil {
			return nil, err
		}
	}

	return proto.Marshal(&ExceptionBlock{
		Transactions: data,
		PreBlock:     pre,
		PostBlock:    post,
	})
}

// EncodeExceptionTx converts ExceptionTx struct into protobuf-encoded ExceptionTx
func EncodeExceptionTx(tx *substate.ExceptionTx) (*ExceptionTx, error) {
	var pre *Alloc
	var err error
	if tx.PreTransaction != nil {
		pre, err = toProtobufAlloc(*tx.PreTransaction)
		if err != nil {
			return nil, err
		}
	}
	var post *Alloc
	if tx.PostTransaction != nil {
		post, err = toProtobufAlloc(*tx.PostTransaction)
		if err != nil {
			return nil, err
		}
	}
	return &ExceptionTx{
		PreTransaction:  pre,
		PostTransaction: post,
		VmException:     &tx.VmException,
	}, nil
}
