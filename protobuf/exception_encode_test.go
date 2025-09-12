package protobuf

import (
	"testing"

	"reflect"

	"github.com/0xsoniclabs/substate/substate"
	"google.golang.org/protobuf/proto"
)

func TestEncodeDecodeExceptionBlock(t *testing.T) {
	// Prepare test data
	preBlock := &substate.WorldState{}
	postBlock := &substate.WorldState{}
	exceptionTx := substate.ExceptionTx{
		PreTransaction:  preBlock,
		PostTransaction: postBlock,
		VmException:     true,
	}
	exceptionBlock := &substate.ExceptionBlock{
		Transactions: map[int]substate.ExceptionTx{1: exceptionTx},
		PreBlock:     preBlock,
		PostBlock:    postBlock,
	}

	// Encode
	encoded, err := EncodeExceptionBlock(exceptionBlock)
	if err != nil {
		t.Fatalf("EncodeExceptionBlock failed: %v", err)
	}

	// Decode
	var pb ExceptionBlock
	if err := proto.Unmarshal(encoded, &pb); err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}
	decoded, err := pb.Decode(nil)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Compare
	if !reflect.DeepEqual(exceptionBlock.Transactions[1].VmException, decoded.Transactions[1].VmException) {
		t.Errorf("VmException mismatch: got %v, want %v", decoded.Transactions[1].VmException, exceptionBlock.Transactions[1].VmException)
	}
}
