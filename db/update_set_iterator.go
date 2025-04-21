package db

import (
	"encoding/binary"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/0xsoniclabs/substate/updateset"
)

func newUpdateSetIterator(db *updateDB, start, end uint64) *updateSetIterator {
	num := make([]byte, 4)
	binary.BigEndian.PutUint32(num, uint32(start))

	r := util.BytesPrefix([]byte(UpdateDBPrefix))
	r.Start = append(r.Start, num...)

	return &updateSetIterator{
		genericIterator: newIterator[*updateset.UpdateSet](db.newIterator(r)),
		db:              db,
		endBlock:        end,
	}
}

type updateSetIterator struct {
	genericIterator[*updateset.UpdateSet]
	db       UpdateDB
	endBlock uint64
}

func (i *updateSetIterator) decode(data rawEntry) (*updateset.UpdateSet, error) {
	key := data.key
	value := data.value

	block, err := DecodeUpdateSetKey(data.key)
	if err != nil {
		return nil, fmt.Errorf("substate: invalid update-set key found: %v - issue: %w", key, err)
	}

	updateSetRLP, err := updateset.DecodeUpdateSetPB(value)
	if err != nil {
		return nil, err
	}

	updateSet, err := updateSetRLP.ToWorldState(i.db.GetCode, block)
	if err != nil {
		return nil, err

	}

	return &updateset.UpdateSet{
		Block:           block,
		WorldState:      updateSet.WorldState,
		DeletedAccounts: updateSetRLP.DeletedAccounts,
	}, nil
}

func (i *updateSetIterator) start(_ int) {
	i.wg.Add(1)

	go func() {
		defer func() {
			close(i.resultCh)
			i.wg.Done()
		}()
		for {
			if !i.iter.Next() {
				return
			}

			key := make([]byte, len(i.iter.Key()))
			copy(key, i.iter.Key())

			// Decode key, if past the end block, stop here.
			// This avoids filling channels which huge data objects that are not consumed.
			block, err := DecodeUpdateSetKey(key)
			if err != nil {
				i.err = err
				return
			}
			if block > i.endBlock {
				return
			}

			value := make([]byte, len(i.iter.Value()))
			copy(value, i.iter.Value())

			raw := rawEntry{key, value}

			us, err := i.decode(raw)
			if err != nil {
				i.err = err
				return
			}

			i.resultCh <- us
		}
	}()
}
