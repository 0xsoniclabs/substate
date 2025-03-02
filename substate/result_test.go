package substate

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/0xsoniclabs/substate/types"
)

func TestAccount_EqualStatus(t *testing.T) {
	res := &Result{Status: 0}
	comparedRes := &Result{Status: 1}

	if res.Equal(comparedRes) {
		t.Fatal("results status are different but equal returned true")
	}

	comparedRes.Status = res.Status
	if !res.Equal(comparedRes) {
		t.Fatal("results status are same but equal returned false")
	}
}

func TestAccount_EqualBloom(t *testing.T) {
	res := &Result{Bloom: types.Bloom{0}}
	comparedRes := &Result{Bloom: types.Bloom{1}}

	if res.Equal(comparedRes) {
		t.Fatal("results Bloom are different but equal returned true")
	}

	comparedRes.Bloom = res.Bloom
	if !res.Equal(comparedRes) {
		t.Fatal("results Bloom are same but equal returned false")
	}
}

func TestAccount_EqualLogs(t *testing.T) {
	res := &Result{Logs: []*types.Log{{Address: types.Address{0}}}}
	comparedRes := &Result{Logs: []*types.Log{{Address: types.Address{1}}}}

	if res.Equal(comparedRes) {
		t.Fatal("results Log are different but equal returned true")
	}

	comparedRes.Logs = res.Logs
	if !res.Equal(comparedRes) {
		t.Fatal("results Log are same but equal returned false")
	}
}

func TestAccount_EqualContractAddress(t *testing.T) {
	res := &Result{ContractAddress: types.Address{0}}
	comparedRes := &Result{ContractAddress: types.Address{1}}

	if res.Equal(comparedRes) {
		t.Fatal("results ContractAddress are different but equal returned true")
	}

	comparedRes.ContractAddress = res.ContractAddress
	if !res.Equal(comparedRes) {
		t.Fatal("results ContractAddress are same but equal returned false")
	}
}

func TestAccount_EqualGasUsed(t *testing.T) {
	res := &Result{GasUsed: 0}
	comparedRes := &Result{GasUsed: 1}

	if res.Equal(comparedRes) {
		t.Fatal("results GasUsed are different but equal returned true")
	}

	comparedRes.GasUsed = res.GasUsed
	if !res.Equal(comparedRes) {
		t.Fatal("results GasUsed are same but equal returned false")
	}
}

func TestResult_NewResult(t *testing.T) {
	status := uint64(0)
	bloom := types.Bloom{0}
	logs := []*types.Log{{Address: types.Address{0}}}
	contractAddress := types.Address{0}
	gasUsed := uint64(0)

	res := NewResult(status, bloom, logs, contractAddress, gasUsed)

	assert.Equal(t, res.Status, status)
	assert.Equal(t, res.Bloom, bloom)
	assert.Equal(t, res.Logs, logs)
	assert.Equal(t, res.ContractAddress, contractAddress)
	assert.Equal(t, res.GasUsed, gasUsed)
}

func TestResult_EqualTopic(t *testing.T) {
	res := &Result{Logs: []*types.Log{{Topics: []types.Hash{{0}}}}}
	comparedRes := &Result{Logs: []*types.Log{{Topics: []types.Hash{{1}}}}}

	assert.Equal(t, false, res.Equal(comparedRes))

	comparedRes.Logs = res.Logs
	assert.Equal(t, true, res.Equal(comparedRes))
}

func TestResult_EqualSelf(t *testing.T) {
	res := &Result{Logs: []*types.Log{{Topics: []types.Hash{{0}}}}}

	assert.Equal(t, true, res.Equal(res))
}

func TestResult_EqualNil(t *testing.T) {
	res := &Result{Logs: []*types.Log{{Topics: []types.Hash{{0}}}}}

	assert.Equal(t, false, res.Equal(nil))
}

func TestResult_String(t *testing.T) {
	res := &Result{
		Status:          0,
		Bloom:           types.Bloom{0},
		Logs:            []*types.Log{{Address: types.Address{0}}},
		ContractAddress: types.Address{0},
		GasUsed:         0,
	}
	expectedString := "Status: 0Bloom: \x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00Contract Address: 0x0000000000000000000000000000000000000000Gas Used: 0{0x0000000000000000000000000000000000000000 [] [] 0 0x0000000000000000000000000000000000000000000000000000000000000000 0 0x0000000000000000000000000000000000000000000000000000000000000000 0 false}"
	assert.Equal(t, expectedString, res.String())
}
