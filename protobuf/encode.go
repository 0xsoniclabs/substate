package protobuf

import (
	"fmt"

	"github.com/OxSonicLabs/Substate/substate"
	"github.com/OxSonicLabs/Substate/types"
	"github.com/OxSonicLabs/Substate/types/hash"
	"google.golang.org/protobuf/proto"
)

// Encode converts aida-substate into protobuf-encoded message
func Encode(ss *substate.Substate, block uint64, tx int) ([]byte, error) {
	bytes, err := proto.Marshal(toProtobufSubstate(ss))
	if err != nil {
		return nil, fmt.Errorf("cannot encode substate into protobuf block: %v,tx %v; %w", block, tx, err)
	}

	return bytes, nil
}

func toProtobufSubstate(ss *substate.Substate) *Substate {
	return &Substate{
		InputAlloc:  toProtobufAlloc(ss.InputSubstate),
		OutputAlloc: toProtobufAlloc(ss.OutputSubstate),
		BlockEnv:    toProtobufBlockEnv(ss.Env),
		TxMessage:   toProtobufTxMessage(ss.Message),
		Result:      toProtobufResult(ss.Result),
	}
}

// toProtobufAlloc converts substate.WorldState into protobuf-encoded Substate_Alloc
func toProtobufAlloc(sw substate.WorldState) *Substate_Alloc {
	world := make([]*Substate_AllocEntry, 0, len(sw))
	for addr, acct := range sw {
		storage := make([]*Substate_Account_StorageEntry, 0, len(acct.Storage))
		for key, value := range acct.Storage {
			storage = append(storage, &Substate_Account_StorageEntry{
				Key:   key.Bytes(),
				Value: value.Bytes(),
			})
		}

		world = append(world, &Substate_AllocEntry{
			Address: addr.Bytes(),
			Account: &Substate_Account{
				Nonce:   &acct.Nonce,
				Balance: acct.Balance.Bytes(),
				Storage: storage,
				Contract: &Substate_Account_CodeHash{
					CodeHash: hash.Keccak256Hash(acct.Code).Bytes(),
				},
			},
		})
	}

	return &Substate_Alloc{Alloc: world}
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
		Difficulty:  se.Difficulty.Bytes(),
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
func toProtobufTxMessage(sm *substate.Message) *Substate_TxMessage {
	dt := Substate_TxMessage_TXTYPE_LEGACY
	txType := &dt
	if sm.ProtobufTxType != nil {
		t := Substate_TxMessage_TxType(*sm.ProtobufTxType)
		txType = &t
	}

	accessList := make([]*Substate_TxMessage_AccessListEntry, len(sm.AccessList))
	for i, entry := range sm.AccessList {
		accessList[i].toProtobufAccessListEntry(&entry)
	}

	blobHashes := make([][]byte, len(sm.BlobHashes))
	for i, hash := range sm.BlobHashes {
		blobHashes[i] = hash.Bytes()
	}

	return &Substate_TxMessage{
		Nonce:         &sm.Nonce,
		GasPrice:      sm.GasPrice.Bytes(),
		Gas:           &sm.Gas,
		From:          sm.From.Bytes(),
		To:            AddressToWrapperspbBytes(sm.To),
		Value:         sm.Value.Bytes(),
		Input:         &Substate_TxMessage_Data{Data: sm.Data},
		TxType:        txType,
		AccessList:    accessList,
		GasFeeCap:     BigIntToWrapperspbBytes(sm.GasFeeCap),
		GasTipCap:     BigIntToWrapperspbBytes(sm.GasTipCap),
		BlobGasFeeCap: BigIntToWrapperspbBytes(sm.BlobGasFeeCap),
		BlobHashes:    blobHashes,
	}
}

// toProtobufAccessListEntry converts types.AccessTuple into protobuf-encoded Substate_TxMessage_AccessListEntry
func (entry *Substate_TxMessage_AccessListEntry) toProtobufAccessListEntry(sat *types.AccessTuple) {
	keys := make([][]byte, len(sat.StorageKeys))
	for i, key := range sat.StorageKeys {
		keys[i] = key.Bytes()
	}

	entry = &Substate_TxMessage_AccessListEntry{
		Address:     sat.Address.Bytes(),
		StorageKeys: keys,
	}
}

// encode converts substate.Results into protobuf-encoded Substate_Result
func toProtobufResult(sr *substate.Result) *Substate_Result {
	logs := make([]*Substate_Result_Log, len(sr.Logs))
	for i, log := range sr.Logs {
		logs[i].toProtobufLog(log)
	}

	return &Substate_Result{
		Status:  &sr.Status,
		Bloom:   sr.Bloom.Bytes(),
		Logs:    logs,
		GasUsed: &sr.GasUsed,
	}
}

// toProtobufLog converts types.Log into protobuf-encoded Substate_Result_log
func (log *Substate_Result_Log) toProtobufLog(sl *types.Log) {
	topics := make([][]byte, len(sl.Topics))
	for i, topic := range sl.Topics {
		topics[i] = topic.Bytes()
	}

	log = &Substate_Result_Log{
		Address: sl.Address.Bytes(),
		Topics:  topics,
		Data:    sl.Data,
	}
}
