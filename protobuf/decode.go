package protobuf

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/OxSonicLabs/Substate/substate"
	"github.com/OxSonicLabs/Substate/types"
	"github.com/OxSonicLabs/Substate/types/hash"
	trlp "github.com/OxSonicLabs/Substate/types/rlp"
	"github.com/syndtr/goleveldb/leveldb"
)

type getCodeFunc = func(types.Hash) ([]byte, error)

// Decode converts protobuf-encoded bytes into aida substate
func (s *Substate) Decode(lookup getCodeFunc, block uint64, tx int) (*substate.Substate, error) {
	input, err := s.GetInputAlloc().decode(lookup)
	if err != nil {
		return nil, err
	}

	output, err := s.GetOutputAlloc().decode(lookup)
	if err != nil {
		return nil, err
	}

	environment := s.GetBlockEnv().decode()

	message, err := s.GetTxMessage().decode(lookup)
	if err != nil {
		return nil, err
	}

	contractAddress := s.GetTxMessage().getContractAddress()
	result := s.GetResult().decode(contractAddress)

	return &substate.Substate{
		InputSubstate:  *input,
		OutputSubstate: *output,
		Env:            environment,
		Message:        message,
		Result:         result,
		Block:          block,
		Transaction:    tx,
	}, nil
}

// decode converts protobuf-encoded Substate_Alloc into aida-comprehensible WorldState
func (alloc *Substate_Alloc) decode(lookup getCodeFunc) (*substate.WorldState, error) {
	world := make(substate.WorldState, len(alloc.GetAlloc()))

	for _, entry := range alloc.GetAlloc() {
		addr, acct := entry.decode()
		address := types.BytesToAddress(addr)
		nonce, balance, _, codehash := acct.decode()

		code, err := lookup(codehash)
		if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
			return nil, fmt.Errorf("Error looking up codehash; %w", err)
		}

		world[address] = substate.NewAccount(nonce, balance, code)
		for _, storage := range acct.GetStorage() {
			key, value := storage.decode()
			world[address].Storage[key] = value
		}
	}

	return &world, nil
}

func (entry *Substate_AllocEntry) decode() ([]byte, *Substate_Account) {
	return entry.GetAddress(), entry.GetAccount()
}

func (acct *Substate_Account) decode() (uint64, *big.Int, []byte, types.Hash) {
	return acct.GetNonce(),
		BytesToBigInt(acct.GetBalance()),
		acct.GetCode(),
		types.BytesToHash(acct.GetCodeHash())
}

func (entry *Substate_Account_StorageEntry) decode() (types.Hash, types.Hash) {
	return types.BytesToHash(entry.GetKey()), types.BytesToHash(entry.GetValue())
}

// decode converts protobuf-encoded Substate_BlockEnv into aida-comprehensible Env
func (env *Substate_BlockEnv) decode() *substate.Env {
	var blockHashes map[uint64]types.Hash = nil
	if env.GetBlockHashes() != nil {
		blockHashes = make(map[uint64]types.Hash, len(env.GetBlockHashes()))
		for _, entry := range env.GetBlockHashes() {
			key, value := entry.decode()
			blockHashes[key] = types.BytesToHash(value)
		}
	}

	return &substate.Env{
		Coinbase:    types.BytesToAddress(env.GetCoinbase()),
		Difficulty:  BytesToBigInt(env.GetDifficulty()),
		GasLimit:    env.GetGasLimit(),
		Number:      env.GetNumber(),
		Timestamp:   env.GetTimestamp(),
		BlockHashes: blockHashes,
		BaseFee:     BytesValueToBigInt(env.GetBaseFee()),
		Random:      BytesValueToHash(env.GetRandom()),
		BlobBaseFee: BytesValueToBigInt(env.GetBlobBaseFee()),
	}
}

func (entry *Substate_BlockEnv_BlockHashEntry) decode() (uint64, []byte) {
	return entry.GetKey(), entry.GetValue()
}

// decode converts protobuf-encoded Substate_TxMessage into aida-comprehensible Message
func (msg *Substate_TxMessage) decode(lookup getCodeFunc) (*substate.Message, error) {

	// to=nil means contract creation
	var pTo *types.Address = nil
	to := msg.GetTo()
	if to != nil {
		address := types.BytesToAddress(to.GetValue())
		pTo = &address
	}

	// In normal cases, pass data directly.
	// In case of contract creation, lookup msg.GetInitCodeHash() and pass that instead
	var data []byte = msg.GetData()
	if pTo == nil {
		code, err := lookup(types.BytesToHash(msg.GetInitCodeHash()))
		if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
			return nil, fmt.Errorf("failed to decode tx message; %w", err)
		}
		data = code
	}

	txType := msg.GetTxType()

	// Berlin hard fork, EIP-2930: Optional access lists
	var accessList types.AccessList = []types.AccessTuple{}
	switch txType {
	case Substate_TxMessage_TXTYPE_ACCESSLIST,
		Substate_TxMessage_TXTYPE_DYNAMICFEE,
		Substate_TxMessage_TXTYPE_BLOB:

		accessList = make([]types.AccessTuple, len(msg.GetAccessList()))
		for i, entry := range msg.GetAccessList() {
			addr, keys := entry.decode()

			address := types.BytesToAddress(addr)
			storageKeys := make([]types.Hash, len(keys))
			for j, key := range keys {
				storageKeys[j] = types.BytesToHash(key)
			}

			accessList[i] = types.AccessTuple{
				Address:     address,
				StorageKeys: storageKeys,
			}
		}
	}

	// London hard fork, EIP-1559: Fee market
	var gasFeeCap *big.Int = BytesToBigInt(msg.GetGasPrice())
	var gasTipCap *big.Int = BytesToBigInt(msg.GetGasPrice())
	switch txType {
	case Substate_TxMessage_TXTYPE_DYNAMICFEE,
		Substate_TxMessage_TXTYPE_BLOB:

		gasFeeCap = BytesValueToBigInt(msg.GetGasFeeCap())
		gasTipCap = BytesValueToBigInt(msg.GetGasTipCap())
	}

	// Cancun hard fork, EIP-4844
	var blobHashes []types.Hash = nil
	switch txType {
	case Substate_TxMessage_TXTYPE_BLOB:
		msgBlobHashes := msg.GetBlobHashes()
		if msgBlobHashes == nil {
			break
		}

		blobHashes = make([]types.Hash, len(msgBlobHashes))
		for i, hash := range msgBlobHashes {
			blobHashes[i] = types.BytesToHash(hash)
		}
	}

	return &substate.Message{
		Nonce:         msg.GetNonce(),
		CheckNonce:    true,
		GasPrice:      BytesToBigInt(msg.GetGasPrice()),
		Gas:           msg.GetGas(),
		From:          types.BytesToAddress(msg.GetFrom()),
		To:            pTo,
		Value:         BytesToBigInt(msg.GetValue()),
		Data:          data,
		AccessList:    accessList,
		GasFeeCap:     gasFeeCap,
		GasTipCap:     gasTipCap,
		BlobGasFeeCap: BytesValueToBigInt(msg.GetBlobGasFeeCap()),
		BlobHashes:    blobHashes,
	}, nil
}

func (entry *Substate_TxMessage_AccessListEntry) decode() ([]byte, [][]byte) {
	return entry.GetAddress(), entry.GetStorageKeys()
}

// getContractAddress returns, the address.Bytes() of the newly created contract,
// returns nil if no contract is created.
func (msg *Substate_TxMessage) getContractAddress() types.Address {
	// *to==nil means no contract creation
	if msg.GetTo() != nil {
		return types.Address{}
	}

	return createAddress(types.BytesToAddress(msg.GetFrom()), msg.GetNonce())
}

// createAddress creates an address given the bytes and the nonce
// mimics crypto.CreateAddress, to avoid cyclical dependency.
func createAddress(addr types.Address, nonce uint64) types.Address {
	data, _ := trlp.EncodeToBytes([]interface{}{addr, nonce})
	return types.BytesToAddress(hash.Keccak256Hash(data).Bytes()[12:])
}

// decode converts protobuf-encoded Substate_Result into aida-comprehensible Result
func (res *Substate_Result) decode(contractAddress types.Address) *substate.Result {
	logs := make([]*types.Log, len(res.GetLogs()))
	for i, log := range res.GetLogs() {
		logs[i] = log.decode()
	}

	return &substate.Result{
		Status:          res.GetStatus(),
		Bloom:           types.BytesToBloom(res.Bloom),
		Logs:            logs,
		ContractAddress: contractAddress,
		GasUsed:         res.GetGasUsed(),
	}
}

func (log *Substate_Result_Log) decode() *types.Log {
	topics := make([]types.Hash, len(log.GetTopics()))
	for i, topic := range log.GetTopics() {
		topics[i] = types.BytesToHash(topic)
	}

	return &types.Log{
		Address: types.BytesToAddress(log.GetAddress()),
		Topics:  topics,
		Data:    log.GetData(),
	}
}
