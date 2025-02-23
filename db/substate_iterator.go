package db

import (
	"fmt"
	"github.com/0xsoniclabs/substate/substate"
	"sync/atomic"
)

func newSubstateIterator(db ISubstateDB, start []byte) *substateIterator {
	return &substateIterator{
		genericIterator: newIterator[*substate.Substate](db.NewIterator([]byte(SubstateDBPrefix), start)),
		db:              db,
	}
}

type substateIterator struct {
	genericIterator[*substate.Substate]
	db ISubstateDB
}

func (i *substateIterator) decode(data rawEntry) (*substate.Substate, error) {
	key := data.key
	value := data.value

	block, tx, err := DecodeSubstateDBKey(data.key)
	if err != nil {
		return nil, fmt.Errorf("invalid substate key: %v; %w", key, err)
	}

	return i.db.decodeToSubstate(value, block, tx)
}

func (i *substateIterator) start(numWorkers int) {
	// Create channels
	errCh := make(chan error, numWorkers)
	rawDataChs := make([]chan rawEntry, numWorkers)
	resultChs := make([]chan *substate.Substate, numWorkers)

	for i := 0; i < numWorkers; i++ {
		rawDataChs[i] = make(chan rawEntry, 10)
		resultChs[i] = make(chan *substate.Substate, 10)
	}

	// Start i => raw data stage
	i.wg.Add(1)
	go func() {
		defer func() {
			for _, c := range rawDataChs {
				close(c)
			}
			i.wg.Done()
		}()
		step := 0
		for i.iter.Next() {
			key := make([]byte, len(i.iter.Key()))
			copy(key, i.iter.Key())
			value := make([]byte, len(i.iter.Value()))
			copy(value, i.iter.Value())

			res := rawEntry{key, value}

			select {
			case <-i.stopCh:
				return
			case rawDataChs[step] <- res: // fall-through
			}
			step = (step + 1) % numWorkers
		}
	}()

	// Start raw data => parsed transaction stage (parallel)
	var numWorkerRunning atomic.Int32
	for w := 0; w < numWorkers; w++ {
		i.wg.Add(1)
		id := w
		numWorkerRunning.Add(1)

		go func() {
			defer func() {
				close(resultChs[id])
				i.wg.Done()
				numWorkerRunning.Add(-1)
				if numWorkerRunning.Load() == 0 {
					close(errCh)
				}
			}()
			for {
				select {
				case <-i.stopCh:
					return
				case raw, ok := <-rawDataChs[id]:
					if !ok {
						return
					}
					transaction, err := i.decode(raw)
					if err != nil {
						errCh <- err
						return
					}
					select {
					case resultChs[id] <- transaction:
					case <-i.stopCh:
						return

					}
				}

			}
		}()
	}

	// Start the go routine moving transactions from parsers to sink in order
	i.wg.Add(1)
	go func() {
		defer func() {
			close(i.resultCh)
			i.wg.Done()
		}()
		step := 0
		for openProducers := numWorkers; openProducers > 0; {
			next, ok := <-resultChs[step%numWorkers]
			if !ok {
				i.err = <-errCh
				return
			}
			if next != nil {
				select {
				case <-i.stopCh:
					return
				case i.resultCh <- next:
				}
			} else {
				openProducers--
			}
			step++
		}
	}()
}
