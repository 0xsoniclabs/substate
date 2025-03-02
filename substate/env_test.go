package substate

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/0xsoniclabs/substate/types"
)

func TestEnv_EqualCoinbase(t *testing.T) {
	env := &Env{
		Coinbase: types.Address{0},
	}
	comparedEnv := &Env{
		Coinbase: types.Address{1},
	}

	if env.Equal(comparedEnv) {
		t.Fatal("envs coinbase are different but equal returned true")
	}

	comparedEnv.Coinbase = env.Coinbase
	if !env.Equal(comparedEnv) {
		t.Fatal("envs coinbase are same but equal returned false")
	}
}

func TestEnv_EqualDifficulty(t *testing.T) {
	env := &Env{
		Difficulty: new(big.Int).SetUint64(0),
	}
	comparedEnv := &Env{
		Difficulty: new(big.Int).SetUint64(1),
	}

	if env.Equal(comparedEnv) {
		t.Fatal("envs difficulty are different but equal returned true")
	}

	comparedEnv.Difficulty = env.Difficulty
	if !env.Equal(comparedEnv) {
		t.Fatal("envs difficulty are same but equal returned false")
	}
}

func TestEnv_EqualGasLimit(t *testing.T) {
	env := &Env{
		GasLimit: 0,
	}
	comparedEnv := &Env{
		GasLimit: 1,
	}

	if env.Equal(comparedEnv) {
		t.Fatal("envs gasLimit are different but equal returned true")
	}

	comparedEnv.GasLimit = env.GasLimit
	if !env.Equal(comparedEnv) {
		t.Fatal("envs gasLimit are same but equal returned false")
	}
}

func TestEnv_EqualNumber(t *testing.T) {
	env := &Env{
		Number: 0,
	}
	comparedEnv := &Env{
		Number: 1,
	}

	if env.Equal(comparedEnv) {
		t.Fatal("envs number are different but equal returned true")
	}

	comparedEnv.Number = env.Number
	if !env.Equal(comparedEnv) {
		t.Fatal("envs number are same but equal returned false")
	}
}

func TestEnv_EqualBlockHashes(t *testing.T) {
	env := &Env{
		BlockHashes: map[uint64]types.Hash{0: types.BytesToHash([]byte{0})},
	}
	comparedEnv := &Env{
		BlockHashes: map[uint64]types.Hash{0: types.BytesToHash([]byte{1})},
	}

	if env.Equal(comparedEnv) {
		t.Fatal("envs hashes for block 0 are different but equal returned true")
	}

	comparedEnv.BlockHashes = map[uint64]types.Hash{1: types.BytesToHash([]byte{1})}

	if env.Equal(comparedEnv) {
		t.Fatal("envs blockHashes are different but equal returned true")
	}

	comparedEnv.BlockHashes = env.BlockHashes
	if !env.Equal(comparedEnv) {
		t.Fatal("envs number are same but equal returned false")
	}
}

func TestEnv_EqualBaseFee(t *testing.T) {
	env := &Env{
		BaseFee: new(big.Int).SetUint64(0),
	}
	comparedEnv := &Env{
		BaseFee: new(big.Int).SetUint64(1),
	}

	if env.Equal(comparedEnv) {
		t.Fatal("envs BaseFee are different but equal returned true")
	}

	comparedEnv.BaseFee = env.BaseFee
	if !env.Equal(comparedEnv) {
		t.Fatal("envs BaseFee are same but equal returned false")
	}
}

func TestEnv_EqualBlobBaseFee(t *testing.T) {
	env := &Env{
		BlobBaseFee: new(big.Int).SetUint64(0),
	}
	comparedEnv := &Env{
		BlobBaseFee: new(big.Int).SetUint64(1),
	}

	if env.Equal(comparedEnv) {
		t.Fatal("envs BlobBaseFee are different but equal returned true")
	}

	comparedEnv.BlobBaseFee = env.BlobBaseFee
	if !env.Equal(comparedEnv) {
		t.Fatal("envs BlobBaseFee are same but equal returned false")
	}
}

func TestEnv_EqualRandom(t *testing.T) {
	env := &Env{
		Random: &types.Hash{0},
	}
	comparedEnv := &Env{
		Random: &types.Hash{1},
	}

	if env.Equal(comparedEnv) {
		t.Fatal("envs Random are different but equal returned true")
	}

	comparedEnv.Random = env.Random
	if !env.Equal(comparedEnv) {
		t.Fatal("envs Random are same but equal returned false")
	}
}

func TestEnv_NewEnv(t *testing.T) {
	coinbase := types.Address{0}
	difficulty := new(big.Int).SetUint64(0)
	gasLimit := uint64(0)
	number := uint64(0)
	timestamp := uint64(0)
	baseFee := new(big.Int).SetUint64(0)
	blobBaseFee := new(big.Int).SetUint64(0)
	blockHashes := map[uint64]types.Hash{0: types.BytesToHash([]byte{0})}
	random := &types.Hash{0}

	env := NewEnv(
		coinbase,
		difficulty,
		gasLimit,
		number,
		timestamp,
		baseFee,
		blobBaseFee,
		blockHashes,
		random,
	)

	assert.Equal(t, coinbase, env.Coinbase)
	assert.Equal(t, 0, env.Difficulty.Cmp(difficulty))
	assert.Equal(t, gasLimit, env.GasLimit)
	assert.Equal(t, number, env.Number)
	assert.Equal(t, timestamp, env.Timestamp)
	assert.Equal(t, len(blockHashes), len(env.BlockHashes))
	assert.Equal(t, 0, env.BaseFee.Cmp(baseFee))
	assert.Equal(t, 0, env.BlobBaseFee.Cmp(blobBaseFee))
	assert.Equal(t, random.String(), env.Random.String())

	for k, v := range blockHashes {
		assert.Equal(t, v, env.BlockHashes[k])
	}
}

func TestEnv_EqualSelf(t *testing.T) {
	env := &Env{
		Coinbase:    types.Address{0},
		Difficulty:  new(big.Int).SetUint64(0),
		GasLimit:    uint64(0),
		Number:      uint64(0),
		Timestamp:   uint64(0),
		BlockHashes: map[uint64]types.Hash{0: types.BytesToHash([]byte{0})},
		BaseFee:     new(big.Int).SetUint64(0),
	}

	assert.True(t, env.Equal(env))
}

func TestEnv_EqualNil(t *testing.T) {
	env := &Env{
		Coinbase:    types.Address{0},
		Difficulty:  new(big.Int).SetUint64(0),
		GasLimit:    uint64(0),
		Number:      uint64(0),
		Timestamp:   uint64(0),
		BlockHashes: map[uint64]types.Hash{0: types.BytesToHash([]byte{0})},
		BaseFee:     new(big.Int).SetUint64(0),
	}

	assert.False(t, env.Equal(nil))
}

func TestEnv_String(t *testing.T) {
	env := &Env{
		Coinbase:    types.Address{0},
		Difficulty:  new(big.Int).SetUint64(0),
		GasLimit:    uint64(0),
		Number:      uint64(0),
		Timestamp:   uint64(0),
		BlockHashes: map[uint64]types.Hash{0: types.BytesToHash([]byte{0})},
		BaseFee:     new(big.Int).SetUint64(0),
	}

	assert.Equal(t, "Coinbase: 0x0000000000000000000000000000000000000000\nDifficulty: 0\nGas Limit: 0\nNumber: 0\nTimestamp: 0\nBase Fee: 0\nBlob Base Fee: <nil>\nBlock Hashes: \nRandom: <nil>\n0: 0x0000000000000000000000000000000000000000000000000000000000000000\n", env.String())
}
