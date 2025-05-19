package utils

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/urfave/cli/v2"
)

// comparePair is a struct that holds two substates to be compared to keep them together
type comparePair struct {
	srcSubstate    *substate.Substate
	targetSubstate *substate.Substate
}

// Compare function compares substates from two databases
func Compare(ctx *cli.Context, src db.SubstateDB, target db.SubstateDB, workers int, first uint64, last uint64) error {
	errChan := make(chan error, 3+workers)
	wg := &sync.WaitGroup{}

	srcSubstateChan := make(chan *substate.Substate, workers*10)
	targetSubstateChan := make(chan *substate.Substate, workers*10)

	compareCtx, cancelCtx := context.WithCancel(ctx.Context)

	wg.Add(1)
	go comparator(compareCtx, srcSubstateChan, targetSubstateChan, errChan, workers, wg)

	// using only src database for substates counter
	var counter uint64

	// start taskpools to retrieve substates
	wg.Add(1)
	go startCompareTaskPool(compareCtx, src, srcSubstateChan, first, last, workers, errChan, ctx, &counter, wg)
	wg.Add(1)
	go startCompareTaskPool(compareCtx, target, targetSubstateChan, first, last, workers, errChan, ctx, nil, wg)

	go func() {
		wg.Wait()
		close(errChan)
	}()

	err, ok := <-errChan
	if ok {
		cancelCtx()
		return err
	}

	total := atomic.LoadUint64(&counter)
	fmt.Printf("%v identical substates were found\n", total)
	cancelCtx()
	return nil
}

func startCompareTaskPool(compareCtx context.Context, dbInstance db.SubstateDB, substateChan chan *substate.Substate, first uint64, last uint64, workers int, errChan chan error, ctx *cli.Context, counter *uint64, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(substateChan)

	feeder := func(block uint64, tx int, substate *substate.Substate, taskPool *db.SubstateTaskPool) error {
		if counter != nil {
			atomic.AddUint64(counter, 1)
		}

		select {
		case <-compareCtx.Done():
		case substateChan <- substate:
		}
		return nil
	}

	var name string
	if counter != nil {
		name += "compare-source"
	} else {
		name += "compare-target"
	}
	taskPool := &db.SubstateTaskPool{
		Name:     name,
		TaskFunc: feeder,

		First: first,
		Last:  last,

		// has to be 1 to keep substates order
		Workers: 1,
		Ctx:     ctx,
		DB:      dbInstance,
	}
	err := taskPool.Execute()
	if err != nil {
		errChan <- err
	}
}

// comparator reads from both channels from the databases the substates and compares them
func comparator(ctx context.Context, srcChan chan *substate.Substate, targetChan chan *substate.Substate, errChan chan error, workers int, wg *sync.WaitGroup) {
	defer wg.Done()
	toCompareChan := make(chan comparePair, workers*10)

	// internal WaitGroup to wait for all workers to finish
	workersWg := &sync.WaitGroup{}

	// launch comparator workers
	for i := 0; i < workers; i++ {
		workersWg.Add(1)
		go func() {
			defer workersWg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case p, ok := <-toCompareChan:
					if !ok {
						return
					}
					err := p.srcSubstate.Equal(p.targetSubstate)
					if err != nil {
						errChan <- err
						return
					}
				}
			}
		}()
	}

	// read from both database channels one value, pair them together and send to workers for comparison
	for {
		// substates are ordered by block and transaction, therefore have to be read by pair
		srcSubstate, ok := <-srcChan
		targetSubstate, ok2 := <-targetChan
		if !ok || !ok2 {
			if ok == ok2 {
				// both channels are closed as same time
				close(toCompareChan)
				break
			}

			// one channel contained additional data
			if ok {
				errChan <- fmt.Errorf("target db doesn't contain substates from %v-%v onwards", srcSubstate.Block, srcSubstate.Transaction)
			} else {
				errChan <- fmt.Errorf("source db doesn't contain substate from %v-%v onwards", targetSubstate.Block, targetSubstate.Transaction)
			}
			close(toCompareChan)
			break
		}

		// pairing substates together
		toCompareChan <- comparePair{srcSubstate: srcSubstate.Clone(), targetSubstate: targetSubstate.Clone()}
	}

	// wait for all workers to finish
	doneChan := make(chan struct{}, 1)
	go func() {
		workersWg.Wait()
		doneChan <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		{
			workersWg.Wait()
		}
	case <-doneChan:
	}
}
