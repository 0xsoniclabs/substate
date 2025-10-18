package protobuf

import (
	"fmt"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/utils"
	"google.golang.org/protobuf/proto"
)

// Encode converts aida-substate into protobuf-encoded message
func Encode(ss *substate.Substate, block uint64, tx int) ([]byte, error) {
	// Field `Account.contract.code_hash` and `TxMessage.input.init_code_hash` are required by the decoder
	// We need to ensure that the code hashes are not nil by calling `HashedCopy` method
	pSubstate, err := toProtobufSubstate(ss)
	if err != nil {
		return nil, err
	}
	h, err := pSubstate.HashedCopy()
	if err != nil {
		return nil, nil
	}
	bytes, err := proto.Marshal(h)
	if err != nil {
		return nil, fmt.Errorf("cannot encode substate into protobuf block: %v,tx %v; %w", block, tx, err)
	}

	return bytes, nil
}

func toProtobufSubstate(ss *substate.Substate) (*Substate, error) {
	input, err := toProtobufAlloc(ss.InputSubstate)
	if err != nil {
		return nil, err
	}
	output, err := toProtobufAlloc(ss.OutputSubstate)
	if err != nil {
		return nil, err
	}
	message, err := toProtobufTxMessage(ss.Message)
	if err != nil {
		return nil, err
	}
	return &Substate{
		InputAlloc:  input,
		OutputAlloc: output,
		BlockEnv:    toProtobufBlockEnv(ss.Env),
		TxMessage:   message,
		Result:      toProtobufResult(ss.Result),
	}, nil
}

// toProtobufAlloc converts substate.WorldState into protobuf-encoded Alloc
func toProtobufAlloc(sw substate.WorldState) (*Alloc, error) {
	world := make([]*AllocEntry, 0, len(sw))
	for addr, acct := range sw {
		storage := make([]*Account_StorageEntry, 0, len(acct.Storage))
		for key, value := range acct.Storage {
			storage = append(storage, &Account_StorageEntry{
				Key:   key.Bytes(),
				Value: value.Bytes(),
			})
		}

		hash, err := utils.Keccak256Hash(acct.Code)
		if err != nil {
			return nil, err
		}
		world = append(world, &AllocEntry{
			Address: addr.Bytes(),
			Account: &Account{
				Nonce:   &acct.Nonce,
				Balance: acct.Balance.Bytes(),
				Storage: storage,
				Contract: &Account_CodeHash{
					CodeHash: hash.Bytes(),
				},
			},
		})
	}

	return &Alloc{Alloc: world}, nil
}

// encode converts substate.Env into protobuf-encoded Substate_BlockEnv
func toProtobufBlockEnv(se *substate.Env) *Substate_BlockEnv {
	blockHashes := make([]*Substate_BlockEnv_BlockHashEntry, 0, len(se.BlockHashes))
	for number, hash := range se.BlockHashes {
		blockHashes = append(blockHashes, &Substate_BlockEnv_BlockHashEntry{
			Key:   &number,
			Value: hash.Bytes(),
		})
	}

	return &Substate_BlockEnv{
		Coinbase:    se.Coinbase.Bytes(),
		Difficulty:  BigIntToBytes(se.Difficulty),
		GasLimit:    &se.GasLimit,
		Number:      &se.Number,
		Timestamp:   &se.Timestamp,
		BlockHashes: blockHashes,
		BaseFee:     BigIntToWrapperspbBytes(se.BaseFee),
		BlobBaseFee: BigIntToWrapperspbBytes(se.BlobBaseFee),
		Random:      HashToWrapperspbBytes(se.Random),
	}
}

// encode converts substate.Message into protobuf-encoded Substate_TxMessage
func toProtobufTxMessage(sm *substate.Message) (*Substate_TxMessage, error) {
	txType := Substate_TxMessage_TXTYPE_LEGACY
	if sm.ProtobufTxType != nil {
		txType = Substate_TxMessage_TxType(*sm.ProtobufTxType)
	}

	accessList := make([]*Substate_TxMessage_AccessListEntry, len(sm.AccessList))
	for i, entry := range sm.AccessList {
		accessList[i] = toProtobufAccessListEntry(&entry)
	}

	blobHashes := make([][]byte, len(sm.BlobHashes))
	for i, hash := range sm.BlobHashes {
		blobHashes[i] = hash.Bytes()
	}

	var txInput isSubstate_TxMessage_Input
	if sm.To == nil {
		hash, err := sm.DataHash()
		if err != nil {
			return nil, err
		}
		txInput = &Substate_TxMessage_InitCodeHash{InitCodeHash: hash.Bytes()}
	} else {
		txInput = &Substate_TxMessage_Data{Data: sm.Data}
	}

	setCodeAuthorizationsList := convertMessageSetCodeAuthorizationToProtobufList(sm)

	return &Substate_TxMessage{
		Nonce:                 &sm.Nonce,
		GasPrice:              BigIntToBytes(sm.GasPrice),
		Gas:                   &sm.Gas,
		From:                  sm.From.Bytes(),
		To:                    AddressToWrapperspbBytes(sm.To),
		Value:                 BigIntToBytes(sm.Value),
		Input:                 txInput,
		TxType:                &txType,
		AccessList:            accessList,
		GasFeeCap:             BigIntToWrapperspbBytes(sm.GasFeeCap),
		GasTipCap:             BigIntToWrapperspbBytes(sm.GasTipCap),
		BlobGasFeeCap:         BigIntToWrapperspbBytes(sm.BlobGasFeeCap),
		BlobHashes:            blobHashes,
		SetCodeAuthorizations: setCodeAuthorizationsList,
	}, nil
}

// convertMessageSetCodeAuthorizationToProtobufList convert substate.Message.SetCodeAuthorization into protobuf-encoded Substate_TxMessage_SetCodeAuthorization
func convertMessageSetCodeAuthorizationToProtobufList(sm *substate.Message) []*Substate_TxMessage_SetCodeAuthorization {
	setCodeAuthorizationsList := make([]*Substate_TxMessage_SetCodeAuthorization, len(sm.SetCodeAuthorizations))
	for i, entry := range sm.SetCodeAuthorizations {
		setCodeAuthorizationsList[i] = &Substate_TxMessage_SetCodeAuthorization{
			ChainId: entry.ChainID.Bytes(),
			Address: entry.Address.Bytes(),
			Nonce:   &entry.Nonce,
			V:       []byte{entry.V},
			R:       entry.R.Bytes(),
			S:       entry.S.Bytes(),
		}
	}
	return setCodeAuthorizationsList
}

// toProtobufAccessListEntry converts types.AccessTuple into protobuf-encoded Substate_TxMessage_AccessListEntry
func toProtobufAccessListEntry(sat *types.AccessTuple) *Substate_TxMessage_AccessListEntry {
	keys := make([][]byte, len(sat.StorageKeys))
	for i, key := range sat.StorageKeys {
		keys[i] = key.Bytes()
	}

	return &Substate_TxMessage_AccessListEntry{
		Address:     sat.Address.Bytes(),
		StorageKeys: keys,
	}
}

// encode converts substate.Results into protobuf-encoded Substate_Result
func toProtobufResult(sr *substate.Result) *Substate_Result {
	logs := make([]*Substate_Result_Log, len(sr.Logs))
	for i, log := range sr.Logs {
		logs[i] = toProtobufLog(log)
	}

	return &Substate_Result{
		Status:  &sr.Status,
		Bloom:   sr.Bloom.Bytes(),
		Logs:    logs,
		GasUsed: &sr.GasUsed,
	}
}

// toProtobufLog converts types.Log into protobuf-encoded Substate_Result_log
func toProtobufLog(sl *types.Log) *Substate_Result_Log {
	topics := make([][]byte, len(sl.Topics))
	for i, topic := range sl.Topics {
		topics[i] = topic.Bytes()
	}

	return &Substate_Result_Log{
		Address: sl.Address.Bytes(),
		Topics:  topics,
		Data:    sl.Data,
	}
}
