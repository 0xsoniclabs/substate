package protobuf

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/types/hash"
	"github.com/holiman/uint256"
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

// decode converts protobuf-encoded Alloc into aida-comprehensible WorldState
func (alloc *Alloc) decode(lookup getCodeFunc) (*substate.WorldState, error) {
	world := make(substate.WorldState, len(alloc.GetAlloc()))

	for _, entry := range alloc.GetAlloc() {
		addr, acct := entry.decode()
		address := types.BytesToAddress(addr)
		nonce, balance, _, codehash := acct.decode()

		code, err := lookup(codehash)
		if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
			return nil, fmt.Errorf("error looking up codehash; %w", err)
		}

		world[address] = substate.NewAccount(nonce, balance, code)
		for _, storage := range acct.GetStorage() {
			key, value := storage.decode()
			world[address].Storage[key] = value
		}
	}

	return &world, nil
}

func (entry *AllocEntry) decode() ([]byte, *Account) {
	return entry.GetAddress(), entry.GetAccount()
}

func (acct *Account) decode() (uint64, *uint256.Int, []byte, types.Hash) {
	return acct.GetNonce(),
		BytesToUint256(acct.GetBalance()),
		acct.GetCode(),
		types.BytesToHash(acct.GetCodeHash())
}

func (entry *Account_StorageEntry) decode() (types.Hash, types.Hash) {
	return types.BytesToHash(entry.GetKey()), types.BytesToHash(entry.GetValue())
}

// decode converts protobuf-encoded BlockEnv into aida-comprehensible Env
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
	var data = msg.GetData()
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
		Substate_TxMessage_TXTYPE_BLOB,
		Substate_TxMessage_TXTYPE_SETCODE:
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
	var gasFeeCap = BytesToBigInt(msg.GetGasPrice())
	var gasTipCap = BytesToBigInt(msg.GetGasPrice())
	switch txType {
	case Substate_TxMessage_TXTYPE_DYNAMICFEE,
		Substate_TxMessage_TXTYPE_BLOB,
		Substate_TxMessage_TXTYPE_SETCODE:
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

	var txTypeInt int32
	switch x := *msg.TxType; x {
	case Substate_TxMessage_TXTYPE_LEGACY:
		txTypeInt = substate.LegacyTxType
	case Substate_TxMessage_TXTYPE_ACCESSLIST:
		txTypeInt = substate.AccessListTxType
	case Substate_TxMessage_TXTYPE_DYNAMICFEE:
		txTypeInt = substate.DynamicFeeTxType
	case Substate_TxMessage_TXTYPE_BLOB:
		txTypeInt = substate.BlobTxType
	case Substate_TxMessage_TXTYPE_SETCODE:
		txTypeInt = substate.SetCodeTxType
	default:
		return nil, fmt.Errorf("unknown tx type: %d", x)
	}

	// Pectra hard fork, EIP-7702
	var setCodeAuthorizationsList []types.SetCodeAuthorization
	switch txType {
	case Substate_TxMessage_TXTYPE_SETCODE:
		setCodeAuthorizationsList = msg.decodeSetCodeAuthorizations()
	}

	return &substate.Message{
		Nonce:                 msg.GetNonce(),
		CheckNonce:            true,
		GasPrice:              BytesToBigInt(msg.GetGasPrice()),
		Gas:                   msg.GetGas(),
		From:                  types.BytesToAddress(msg.GetFrom()),
		To:                    pTo,
		Value:                 BytesToBigInt(msg.GetValue()),
		Data:                  data,
		ProtobufTxType:        &txTypeInt,
		AccessList:            accessList,
		GasFeeCap:             gasFeeCap,
		GasTipCap:             gasTipCap,
		BlobGasFeeCap:         BytesValueToBigInt(msg.GetBlobGasFeeCap()),
		BlobHashes:            blobHashes,
		SetCodeAuthorizations: setCodeAuthorizationsList,
	}, nil
}

// decodeSetCodeAuthorizations retrieves list of SetCodeAuthorizations
func (msg *Substate_TxMessage) decodeSetCodeAuthorizations() []types.SetCodeAuthorization {
	setCodeAuthorizationsList := make([]types.SetCodeAuthorization, len(msg.GetSetCodeAuthorizations()))
	for i, entry := range msg.GetSetCodeAuthorizations() {
		chainId, addr, nonce, v, r, s := entry.decode()
		setCodeAuthorizationsList[i] = types.SetCodeAuthorization{
			ChainID: *BytesToUint256(chainId),
			Address: types.BytesToAddress(addr),
			Nonce:   nonce,
			V:       v[0],
			R:       *BytesToUint256(r),
			S:       *BytesToUint256(s),
		}
	}
	return setCodeAuthorizationsList
}

func (entry *Substate_TxMessage_AccessListEntry) decode() ([]byte, [][]byte) {
	return entry.GetAddress(), entry.GetStorageKeys()
}

func (entry *Substate_TxMessage_SetCodeAuthorization) decode() ([]byte, []byte, uint64, []byte, []byte, []byte) {
	return entry.GetChainId(), entry.GetAddress(), entry.GetNonce(), entry.GetV(), entry.GetR(), entry.GetS()
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
	// This is equivalent to the RLP encoding of []interface{}{addr, nonce}
	// Old code using trlp:
	// data, _ := trlp.EncodeToBytes([]interface{}{addr, nonce})

	// Create a new buffer to build the RLP-like encoding for address creation
	buffer := bytes.NewBuffer([]byte{})

	// Encode the nonce as a byte slice using custom encoding
	nonceBytes := encodeUint(nonce)

	// Calculate the total size of the encoded data:
	// 0xc0 is the RLP list prefix, plus the lengths of nonceBytes, addr, and 1 for the address type
	size := 0xc0 + len(nonceBytes) + len(addr) + 1

	// Write the size byte to the buffer
	buffer.Write([]byte{byte(size)})

	// Write the address type prefix (0x94 for 20-byte address) to the buffer
	buffer.Write([]byte{0x94})

	// Write the address bytes to the buffer
	buffer.Write(addr.Bytes())

	// Write the encoded nonce bytes to the buffer
	buffer.Write(nonceBytes)

	// Get the final byte slice representing the encoded data
	data := buffer.Bytes()

	return types.BytesToAddress(hash.Keccak256Hash(data).Bytes()[12:])
}

// encodeUint encodes a uint64 into a byte array.
func encodeUint(i uint64) []byte {
	if i == 0 {
		return []byte{0x80}
	} else if i < 128 {
		// fits single byte
		return []byte{byte(i)}
	} else {
		// 1 byte header + maximum 8 bytes for uint64
		buf := make([]byte, 9)
		s := putInt(buf[1:], i)
		buf[0] = 0x80 + byte(s)
		return buf[:s+1]
	}
}

// putInt writes i to the beginning of b in big endian byte
// order, using the least number of bytes needed to represent i.
func putInt(b []byte, i uint64) (size int) {
	switch {
	case i < (1 << 8):
		b[0] = byte(i)
		return 1
	case i < (1 << 16):
		b[0] = byte(i >> 8)
		b[1] = byte(i)
		return 2
	case i < (1 << 24):
		b[0] = byte(i >> 16)
		b[1] = byte(i >> 8)
		b[2] = byte(i)
		return 3
	case i < (1 << 32):
		b[0] = byte(i >> 24)
		b[1] = byte(i >> 16)
		b[2] = byte(i >> 8)
		b[3] = byte(i)
		return 4
	case i < (1 << 40):
		b[0] = byte(i >> 32)
		b[1] = byte(i >> 24)
		b[2] = byte(i >> 16)
		b[3] = byte(i >> 8)
		b[4] = byte(i)
		return 5
	case i < (1 << 48):
		b[0] = byte(i >> 40)
		b[1] = byte(i >> 32)
		b[2] = byte(i >> 24)
		b[3] = byte(i >> 16)
		b[4] = byte(i >> 8)
		b[5] = byte(i)
		return 6
	case i < (1 << 56):
		b[0] = byte(i >> 48)
		b[1] = byte(i >> 40)
		b[2] = byte(i >> 32)
		b[3] = byte(i >> 24)
		b[4] = byte(i >> 16)
		b[5] = byte(i >> 8)
		b[6] = byte(i)
		return 7
	default:
		b[0] = byte(i >> 56)
		b[1] = byte(i >> 48)
		b[2] = byte(i >> 40)
		b[3] = byte(i >> 32)
		b[4] = byte(i >> 24)
		b[5] = byte(i >> 16)
		b[6] = byte(i >> 8)
		b[7] = byte(i)
		return 8
	}
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
