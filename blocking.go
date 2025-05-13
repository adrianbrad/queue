package queue

import (
	"encoding/json"
	"sync"
)

var _ Queue[any] = (*Blocking[any])(nil)

// Blocking is a Queue implementation that additionally supports operations
// that wait for the queue to have available items, and wait for a slot to
// become available in case the queue is full.
// ! The Blocking Queue shares most functionality with channels. If you do
// not make use of Peek, Reset and Contains methods you are safe to use channels instead.
//
// It supports operations for retrieving and adding elements to a FIFO queue.
// If there are no elements available the retrieve operations wait until
// elements are added to the queue.
type Blocking[T comparable] struct {
	initialElems []T
	elems        []T
	capacity     *int

	// synchronization
	lock         sync.RWMutex
	notEmptyCond *sync.Cond
	notFullCond  *sync.Cond
}

// NewBlocking returns a new Blocking Queue containing the given elements.
func NewBlocking[T comparable](
	elems []T,
	opts ...Option,
) *Blocking[T] {
	options := options{
		capacity: nil,
	}

	for _, o := range opts {
		o.apply(&options)
	}

	// Store initial elements
	initialElems := make([]T, len(elems))
	copy(initialElems, elems)

	queue := &Blocking[T]{
		elems:        elems,
		initialElems: initialElems,
		capacity:     options.capacity,
		lock:         sync.RWMutex{},
	}

	queue.notEmptyCond = sync.NewCond(&queue.lock)
	queue.notFullCond = sync.NewCond(&queue.lock)

	if queue.capacity != nil {
		if len(queue.elems) > *queue.capacity {
			queue.elems = queue.elems[:*queue.capacity]
		}
	}

	return queue
}

// ==================================Insertion=================================

// OfferWait inserts the element to the tail the queue.
// It waits for necessary space to become available.
func (bq *Blocking[T]) OfferWait(elem T) {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	for bq.isFull() {
		bq.notFullCond.Wait()
	}

	bq.elems = append(bq.elems, elem)

	bq.notEmptyCond.Signal()
}

// Offer inserts the element to the tail the queue.
// If the queue is full it returns the ErrQueueIsFull error.
func (bq *Blocking[T]) Offer(elem T) error {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	if bq.isFull() {
		return ErrQueueIsFull
	}

	bq.elems = append(bq.elems, elem)

	bq.notEmptyCond.Signal()

	return nil
}

// Reset sets the queue to its initial state with the original elements.
func (bq *Blocking[T]) Reset() {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	// Restore initial elements
	bq.elems = make([]T, len(bq.initialElems))
	copy(bq.elems, bq.initialElems)

	bq.notEmptyCond.Broadcast()
}

// ===================================Removal==================================

// GetWait removes and returns the head of the elements queue.
// If no element is available it waits until the queue
// has an element available.
func (bq *Blocking[T]) GetWait() (v T) {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	for bq.isEmpty() {
		bq.notEmptyCond.Wait()
	}

	elem := bq.elems[0]
	bq.elems = bq.elems[1:]

	bq.notFullCond.Signal()

	return elem
}

// Get removes and returns the head of the elements queue.
// If no element is available it returns an ErrNoElementsAvailable error.
func (bq *Blocking[T]) Get() (v T, _ error) {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	return bq.get()
}

// Clear removes and returns all elements from the queue.
func (bq *Blocking[T]) Clear() []T {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	defer bq.notFullCond.Broadcast()

	removed := make([]T, len(bq.elems))
	copy(removed, bq.elems)
	bq.elems = bq.elems[:0]

	return removed
}

// Iterator returns an iterator over the elements in this queue.
// It removes the elements from the queue.
func (bq *Blocking[T]) Iterator() <-chan T {
	bq.lock.RLock()
	defer bq.lock.RUnlock()

	// use a buffered channel to avoid blocking the iterator.
	iteratorCh := make(chan T, bq.size())

	// close the channel when the function returns.
	defer close(iteratorCh)

	// iterate over the elements and send them to the channel.
	for {
		elem, err := bq.get()
		if err != nil {
			break
		}

		iteratorCh <- elem
	}

	return iteratorCh
}

// =================================Examination================================

// Peek retrieves but does not return the head of the queue.
// If no element is available it returns an ErrNoElementsAvailable error.
func (bq *Blocking[T]) Peek() (v T, _ error) {
	bq.lock.RLock()
	defer bq.lock.RUnlock()

	if bq.isEmpty() {
		return v, ErrNoElementsAvailable
	}

	return bq.elems[0], nil
}

// PeekWait retrieves but does not return the head of the queue.
// If no element is available it waits until the queue
// has an element available.
func (bq *Blocking[T]) PeekWait() T {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	for bq.isEmpty() {
		bq.notEmptyCond.Wait()
	}

	elem := bq.elems[0]

	// send the not empty signal again in case any remove method waits.
	bq.notEmptyCond.Signal()

	return elem
}

// Size returns the number of elements in the queue.
func (bq *Blocking[T]) Size() int {
	bq.lock.RLock()
	defer bq.lock.RUnlock()

	return len(bq.elems)
}

// Contains returns true if the queue contains the given element.
func (bq *Blocking[T]) Contains(elem T) bool {
	bq.lock.RLock()
	defer bq.lock.RUnlock()

	for _, e := range bq.elems {
		if e == elem {
			return true
		}
	}

	return false
}

// IsEmpty returns true if the queue is empty.
func (bq *Blocking[T]) IsEmpty() bool {
	bq.lock.RLock()
	defer bq.lock.RUnlock()

	return bq.isEmpty()
}

// ===================================Helpers==================================

// isEmpty returns true if the queue is empty.
func (bq *Blocking[T]) isEmpty() bool {
	return len(bq.elems) == 0
}

// isFull returns true if the queue is full.
func (bq *Blocking[T]) isFull() bool {
	if bq.capacity == nil {
		return false
	}

	return len(bq.elems) >= *bq.capacity
}

func (bq *Blocking[T]) size() int {
	return len(bq.elems)
}

func (bq *Blocking[T]) get() (v T, _ error) {
	if bq.isEmpty() {
		return v, ErrNoElementsAvailable
	}

	elem := bq.elems[0]
	bq.elems = bq.elems[1:]

	bq.notFullCond.Signal()

	return elem, nil
}

// MarshalJSON serializes the Blocking queue to JSON.
func (bq *Blocking[T]) MarshalJSON() ([]byte, error) {
	bq.lock.RLock()
	defer bq.lock.RUnlock()

	if bq.isEmpty() {
		return []byte("[]"), nil
	}

	return json.Marshal(bq.elems)
}
