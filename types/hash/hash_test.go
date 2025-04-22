package hash

import (
	"errors"
	"reflect"
	"testing"

	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestKeccak256Hash_Success(t *testing.T) {
	input := []byte("test")
	actual := Keccak256Hash(input)
	expected := types.Hash{0x9c, 0x22, 0xff, 0x5f, 0x21, 0xf0, 0xb8, 0x1b, 0x11, 0x3e, 0x63, 0xf7, 0xdb, 0x6d, 0xa9, 0x4f, 0xed, 0xef, 0x11, 0xb2, 0x11, 0x9b, 0x40, 0x88, 0xb8, 0x96, 0x64, 0xfb, 0x9a, 0x3c, 0xb6, 0x58}
	assert.Equal(t, expected, actual)
}

func TestKeccak256Hash_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mock DB
	mock := NewMockKeccakState(ctrl)
	newKeccakStateDelegate = func() KeccakState {
		return mock
	}
	mockErr := errors.New("mock error")

	// Test Write error
	func() {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		mock.EXPECT().Write(gomock.Any()).Return(0, mockErr)
		h := Keccak256Hash([]byte("input"))
		assert.Equal(t, types.Hash{}, h)
	}()

	// Test Read error
	func() {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		mock.EXPECT().Write(gomock.Any()).Return(0, nil).AnyTimes()
		mock.EXPECT().Read(gomock.Any()).Return(0, mockErr)
		h := Keccak256Hash([]byte("input"))
		assert.Equal(t, types.Hash{}, h)
	}()
}

func TestMockDelegate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// test mock
	mockFunc := func() KeccakState {
		return nil
	}
	mockNewKeccakStateDelegate(mockFunc)
	assert.Equal(t, reflect.ValueOf(mockFunc).Pointer(), reflect.ValueOf(newKeccakStateDelegate).Pointer())

	// test unmock
	resetNewKeccakStateDelegate()
	assert.Equal(t, reflect.ValueOf(NewKeccakState).Pointer(), reflect.ValueOf(newKeccakStateDelegate).Pointer())
}
