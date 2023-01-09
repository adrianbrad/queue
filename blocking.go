package queue

import (
	"sync"
)

var _ Queue[any] = (*Blocking[any])(nil)

// Blocking provides a read-only queue for a list of T.
//
// It supports operations for retrieving and adding elements to a FIFO queue.
// If there are no elements available the retrieve operations wait until
// elements are added to the queue.
type Blocking[T any] struct {
	// elements queue
	elements      []T
	elementsIndex int

	lock         sync.Mutex
	notEmptyCond *sync.Cond
}

// NewBlocking returns a new Blocking Queue containing the given elements..
func NewBlocking[T any](elements []T) *Blocking[T] {
	b := &Blocking[T]{
		elements:      elements,
		elementsIndex: 0,
		lock:          sync.Mutex{},
	}

	b.notEmptyCond = sync.NewCond(&b.lock)

	return b
}

// Take removes and returns the head of the elements queue.
// If no element is available it waits until the queue
// has an element available.
//
// It does not actually remove elements from the elements slice, but
// it's incrementing the underlying index.
func (q *Blocking[T]) Take() (v T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	idx := q.getNextIndexOrWait()

	elem := q.elements[idx]

	q.elementsIndex++

	return elem
}

// Get removes and returns the head of the elements queue.
// If no element is available it returns an ErrNoElementsAvailable error.
//
// It does not actually remove elements from the elements slice, but
// it's incrementing the underlying index.
func (q *Blocking[T]) Get() (v T, _ error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.elementsIndex >= len(q.elements) {
		return v, ErrNoElementsAvailable
	}

	elem := q.elements[q.elementsIndex]

	q.elementsIndex++

	return elem, nil
}

func (q *Blocking[T]) getNextIndexOrWait() int {
	if q.elementsIndex < len(q.elements) {
		return q.elementsIndex
	}

	q.notEmptyCond.Wait()

	return q.getNextIndexOrWait()
}

// Put inserts the element to the tail the queue,
// while also increasing the queue size.
func (q *Blocking[T]) Put(elem T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.elements = append(q.elements, elem)

	q.notEmptyCond.Signal()
}

// Peek retrieves but does not return the head of the queue.
func (q *Blocking[T]) Peek() T {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.elementsIndex == len(q.elements) {
		q.notEmptyCond.Wait()
	}

	elem := q.elements[q.elementsIndex]

	return elem
}

// Reset sets the queue elements index to 0. The queue will be in its initial
// state.
func (q *Blocking[T]) Reset() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.elementsIndex = 0

	q.notEmptyCond.Broadcast()
}
