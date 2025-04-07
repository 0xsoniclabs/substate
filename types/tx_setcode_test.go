package types

import (
	"testing"

	"github.com/holiman/uint256"
)

func Test_EqualSetCodeAuthorization(t *testing.T) {
	msg := SetCodeAuthorization{ChainID: *uint256.NewInt(1), Address: Address{0}, Nonce: 1, V: 1, R: *uint256.NewInt(1), S: *uint256.NewInt(1)}
	compared := SetCodeAuthorization{ChainID: *uint256.NewInt(2), Address: Address{0}, Nonce: 1, V: 1, R: *uint256.NewInt(1), S: *uint256.NewInt(1)}

	if msg.Equal(compared) {
		t.Fatal("messages setCodeAuthorizations have different chainId but equal returned true")
	}

	compared = SetCodeAuthorization{ChainID: *uint256.NewInt(1), Address: Address{1}, Nonce: 1, V: 1, R: *uint256.NewInt(1), S: *uint256.NewInt(1)}
	if msg.Equal(compared) {
		t.Fatal("messages setCodeAuthorizations have different address but equal returned true")
	}

	compared = SetCodeAuthorization{ChainID: *uint256.NewInt(1), Address: Address{0}, Nonce: 2, V: 1, R: *uint256.NewInt(1), S: *uint256.NewInt(1)}
	if msg.Equal(compared) {
		t.Fatal("messages setCodeAuthorizations have different nonce but equal returned true")
	}

	compared = SetCodeAuthorization{ChainID: *uint256.NewInt(1), Address: Address{0}, Nonce: 1, V: 2, R: *uint256.NewInt(1), S: *uint256.NewInt(1)}
	if msg.Equal(compared) {
		t.Fatal("messages setCodeAuthorizations have different V but equal returned true")
	}

	compared = SetCodeAuthorization{ChainID: *uint256.NewInt(1), Address: Address{0}, Nonce: 1, V: 1, R: *uint256.NewInt(2), S: *uint256.NewInt(1)}
	if msg.Equal(compared) {
		t.Fatal("messages setCodeAuthorizations have different R but equal returned true")
	}

	compared = SetCodeAuthorization{ChainID: *uint256.NewInt(1), Address: Address{0}, Nonce: 1, V: 1, R: *uint256.NewInt(1), S: *uint256.NewInt(2)}
	if msg.Equal(compared) {
		t.Fatal("messages setCodeAuthorizations have different S but equal returned true")
	}

	compared = msg
	if !msg.Equal(compared) {
		t.Fatal("messages setCodeAuthorizations are same but equal returned false")
	}
}
