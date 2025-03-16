package db

import (
	"encoding/binary"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUpdateDB_PutMetadataSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	codeDb := NewMockCodeDB(ctrl)
	db := &updateDB{
		CodeDB: codeDb,
	}

	// Test successful put
	interval := uint64(1000)
	size := uint64(500)

	// Create expected byte slices
	intervalBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(intervalBytes, interval)

	sizeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBytes, size)

	// Set expectations
	codeDb.EXPECT().Put([]byte(UpdatesetIntervalKey), intervalBytes).Return(nil)
	codeDb.EXPECT().Put([]byte(UpdatesetSizeKey), sizeBytes).Return(nil)

	err := db.PutMetadata(interval, size)
	assert.Nil(t, err)
}

func TestUpdateDB_PutMetadataPutIntervalKeyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	codeDb := NewMockCodeDB(ctrl)
	db := &updateDB{
		CodeDB: codeDb,
	}

	interval := uint64(1000)
	size := uint64(500)

	// Test error when putting interval key
	expectedErr := errors.New("interval put error")
	codeDb.EXPECT().Put([]byte(UpdatesetIntervalKey), gomock.Any()).Return(expectedErr)

	err := db.PutMetadata(interval, size)
	assert.Equal(t, expectedErr, err)
}

func TestUpdateDB_PutMetadataPutSizeKeyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	codeDb := NewMockCodeDB(ctrl)
	db := &updateDB{
		CodeDB: codeDb,
	}

	interval := uint64(1000)
	size := uint64(500)

	// Test error when putting size key
	expectedErr := errors.New("size put error")
	codeDb.EXPECT().Put([]byte(UpdatesetIntervalKey), gomock.Any()).Return(nil)
	codeDb.EXPECT().Put([]byte(UpdatesetSizeKey), gomock.Any()).Return(expectedErr)

	err := db.PutMetadata(interval, size)
	assert.Equal(t, expectedErr, err)
}

func TestUpdateDB_GetMetadataSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	codeDb := NewMockCodeDB(ctrl)
	db := &updateDB{
		CodeDB: codeDb,
	}

	// Test successful get
	interval := uint64(1000)
	size := uint64(500)

	// Create byte slices to return
	intervalBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(intervalBytes, interval)

	sizeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBytes, size)

	// Set expectations
	codeDb.EXPECT().Get([]byte(UpdatesetIntervalKey)).Return(intervalBytes, nil)
	codeDb.EXPECT().Get([]byte(UpdatesetSizeKey)).Return(sizeBytes, nil)

	resultInterval, resultSize, err := db.GetMetadata()
	assert.Nil(t, err)
	assert.Equal(t, interval, resultInterval)
	assert.Equal(t, size, resultSize)
}

func TestUpdateDB_GetMetadataIntervalKeyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	codeDb := NewMockCodeDB(ctrl)
	db := &updateDB{
		CodeDB: codeDb,
	}

	// Test error when getting interval
	expectedErr := errors.New("interval get error")
	codeDb.EXPECT().Get([]byte(UpdatesetIntervalKey)).Return(nil, expectedErr)

	resultInterval, resultSize, err := db.GetMetadata()
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, uint64(0), resultInterval)
	assert.Equal(t, uint64(0), resultSize)
}

func TestUpdateDB_GetMetadataSizeKeyFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	codeDb := NewMockCodeDB(ctrl)
	db := &updateDB{
		CodeDB: codeDb,
	}

	// Test error when getting size after successful interval get
	expectedErr := errors.New("size get error")

	// Create byte slice for interval
	intervalBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(intervalBytes, uint64(1000))

	codeDb.EXPECT().Get([]byte(UpdatesetIntervalKey)).Return(intervalBytes, nil)
	codeDb.EXPECT().Get([]byte(UpdatesetSizeKey)).Return(nil, expectedErr)

	resultInterval, resultSize, err := db.GetMetadata()
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, uint64(0), resultInterval)
	assert.Equal(t, uint64(0), resultSize)
}
