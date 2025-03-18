package db

import (
	"encoding/binary"
)

const (
	MetadataPrefix       = "md"
	UpdatesetPrefix      = "us"
	UpdatesetIntervalKey = MetadataPrefix + UpdatesetPrefix + "in"
	UpdatesetSizeKey     = MetadataPrefix + UpdatesetPrefix + "si"
)

// PutMetadata into db
func (db *updateDB) PutMetadata(interval, size uint64) error {

	byteInterval := make([]byte, 8)
	binary.BigEndian.PutUint64(byteInterval, interval)

	if err := db.Put([]byte(UpdatesetIntervalKey), byteInterval); err != nil {
		return err
	}

	sizeInterval := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeInterval, size)

	if err := db.Put([]byte(UpdatesetSizeKey), sizeInterval); err != nil {
		return err
	}

	return nil
}

// GetMetadata from db
func (db *updateDB) GetMetadata() (uint64, uint64, error) {
	byteInterval, err := db.Get([]byte(UpdatesetIntervalKey))
	if err != nil {
		return 0, 0, err
	}

	byteSize, err := db.Get([]byte(UpdatesetSizeKey))
	if err != nil {
		return 0, 0, err
	}

	return binary.BigEndian.Uint64(byteInterval), binary.BigEndian.Uint64(byteSize), nil
}
