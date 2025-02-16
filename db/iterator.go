package db

import (
	"errors"
	"sync"

	ldbiterator "github.com/syndtr/goleveldb/leveldb/iterator"
)

// IIterator iterates over a database's key/value pairs in ascending key order.
//
// When it encounters an error any seek will return false and will yield no key/
// value pairs. The error can be queried by calling the Error method. Calling
// Release is still necessary.
//
// An Iterator must be released after use, but it is not necessary to read an
// Iterator until exhaustion. An Iterator is not safe for concurrent use, but it
// is safe to use multiple iterators concurrently.
type IIterator[T comparable] interface {
	// Next moves the Iterator to the next key/value pair. It returns whether the
	// Iterator is exhausted.
	Next() bool

	// Error returns any accumulated error. Exhausting all the key/value pairs
	// is not considered to be an error.
	Error() error

	// Start starts the iteration process.
	start(numWorkers int)

	// Value returns the current value of type T, or nil if done. The
	// caller should not modify the contents of the returned slice, and its contents
	// may change on the next call to Next.
	Value() T

	// Release releases associated resources. Release should always succeed and can
	// be called multiple times without causing error.
	Release()

	// decode data returned from DB to given type T.
	decode(data rawEntry) (T, error)
}

type rawEntry struct {
	key   []byte
	value []byte
}

type Iterator[T comparable] struct {
	err      error
	iter     ldbiterator.Iterator
	resultCh chan T
	wg       *sync.WaitGroup
	cur      T
	stopCh   chan any
}

func newIterator[T comparable](iter ldbiterator.Iterator) Iterator[T] {
	return Iterator[T]{
		iter:     iter,
		resultCh: make(chan T, 10),
		wg:       new(sync.WaitGroup),
		stopCh:   make(chan any, 1),
	}
}

// Next returns false if Iterator is at its end. Otherwise, it returns true.
// Note: False does not stop the Iterator. Release() should be called.
func (i *Iterator[T]) Next() bool {
	if i.err != nil {
		return false
	}
	i.cur = <-i.resultCh
	var zero T
	return i.cur != zero
}

// Error returns iterators error if any.
func (i *Iterator[T]) Error() error {
	return errors.Join(i.err, i.iter.Error())
}

// Value returns current value hold by the Iterator.
func (i *Iterator[T]) Value() T {
	return i.cur
}

// Release the Iterator and wait until all threads are closed gracefully.
func (i *Iterator[T]) Release() {
	close(i.stopCh)
	i.wg.Wait()
	i.iter.Release()
}

func isNil[T comparable](arg T) bool {
	var t T
	return arg == t
}
