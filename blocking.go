package queue

import (
	"sync"
)

var _ Queue[any] = (*Blocking[any])(nil)

// Blocking is a Queue implementation that additionally supports operations
// that wait for the queue to have available items, and wait for a slot to
// become available in case the queue is full.
// ! The Blocking Queue shares most functionality with channels. If you do
// not make use of Peek or Reset methods you are safe to use channels instead.
//
// It supports operations for retrieving and adding elements to a FIFO queue.
// If there are no elements available the retrieve operations wait until
// elements are added to the queue.
type Blocking[T any] struct {
	// elements queue
	elements      []T
	elementsIndex int

	initialLen int

	capacity *int

	// synchronization
	initCondsOnce sync.Once
	lock          sync.Mutex
	notEmptyCond  *sync.Cond
	notFullCond   *sync.Cond
}

// NewBlocking returns a new Blocking Queue containing the given elements.
func NewBlocking[T any](
	elems []T,
	opts ...Option,
) *Blocking[T] {
	options := options{
		capacity: nil,
	}

	for _, o := range opts {
		o.apply(&options)
	}

	queue := &Blocking[T]{
		elements:      elems,
		elementsIndex: 0,
		initialLen:    len(elems),
		capacity:      options.capacity,
		lock:          sync.Mutex{},
	}

	queue.initConds()

	if queue.capacity != nil {
		if len(queue.elements) > *queue.capacity {
			queue.elements = queue.elements[:*queue.capacity]
		}
	}

	return queue
}

// initConds can only be run once, it enables the Blocking queue to have
// a valid zero value.
func (bq *Blocking[T]) initConds() {
	bq.initCondsOnce.Do(func() {
		bq.notEmptyCond = sync.NewCond(&bq.lock)
		bq.notFullCond = sync.NewCond(&bq.lock)
	})
}

// ==================================Insertion=================================

// OfferWait inserts the element to the tail the queue.
// It waits for necessary space to become available.
func (bq *Blocking[T]) OfferWait(elem T) {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	bq.initConds()

	if bq.isFull() {
		bq.notFullCond.Wait()
	}

	bq.elements = append(bq.elements, elem)

	bq.notEmptyCond.Signal()
}

// Offer inserts the element to the tail the queue.
// If the queue is full it returns the ErrQueueIsFull error.
func (bq *Blocking[T]) Offer(elem T) error {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	bq.initConds()

	if bq.isFull() {
		return ErrQueueIsFull
	}

	bq.elements = append(bq.elements, elem)

	bq.notEmptyCond.Signal()

	return nil
}

// Reset sets the queue elements index to 0. The queue will be in its initial
// state.
func (bq *Blocking[T]) Reset() {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	bq.initConds()

	bq.elementsIndex = 0

	bq.elements = bq.elements[:bq.initialLen]

	bq.notEmptyCond.Broadcast()
}

// ===================================Removal==================================

// GetWait removes and returns the head of the elements queue.
// If no element is available it waits until the queue
// has an element available.
//
// It does not actually remove elements from the elements slice, but
// it's incrementing the underlying index.
func (bq *Blocking[T]) GetWait() (v T) {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	bq.initConds()

	defer bq.notFullCond.Signal()

	idx := bq.getNextIndexOrWait()

	elem := bq.elements[idx]

	bq.elementsIndex++

	return elem
}

// Get removes and returns the head of the elements queue.
// If no element is available it returns an ErrNoElementsAvailable error.
//
// It does not actually remove elements from the elements slice, but
// it's incrementing the underlying index.
func (bq *Blocking[T]) Get() (v T, _ error) {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	bq.initConds()

	defer bq.notFullCond.Signal()

	if bq.isEmpty() {
		return v, ErrNoElementsAvailable
	}

	elem := bq.elements[bq.elementsIndex]

	bq.elementsIndex++

	return elem, nil
}

// =================================Examination================================

// Peek retrieves but does not return the head of the queue.
// If no element is available it returns an ErrNoElementsAvailable error.
func (bq *Blocking[T]) Peek() (v T, _ error) {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	bq.initConds()

	if bq.isEmpty() {
		return v, ErrNoElementsAvailable
	}

	elem := bq.elements[bq.elementsIndex]

	return elem, nil
}

// PeekWait retrieves but does not return the head of the queue.
// If no element is available it waits until the queue
// has an element available.
func (bq *Blocking[T]) PeekWait() T {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	bq.initConds()

	if bq.isEmpty() {
		bq.notEmptyCond.Wait()
	}

	elem := bq.elements[bq.elementsIndex]

	// send the not empty signal again in case any remove method waits.
	bq.notEmptyCond.Signal()

	return elem
}

// Size returns the number of elements in the queue.
func (bq *Blocking[T]) Size() int {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	return len(bq.elements) - bq.elementsIndex
}

// ===================================Helpers==================================

func (bq *Blocking[T]) getNextIndexOrWait() int {
	if !bq.isEmpty() {
		return bq.elementsIndex
	}

	bq.notEmptyCond.Wait()

	return bq.getNextIndexOrWait()
}

func (bq *Blocking[T]) isEmpty() bool {
	return bq.elementsIndex >= len(bq.elements)
}

func (bq *Blocking[T]) isFull() bool {
	if bq.capacity == nil {
		return false
	}

	return len(bq.elements)-bq.elementsIndex >= *bq.capacity
}
