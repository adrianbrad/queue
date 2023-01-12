package queue

import (
	"fmt"
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

	// in case capacity is provided
	initialLen *int
	capacity   *int

	// synchronization
	lock         sync.Mutex
	notEmptyCond *sync.Cond
	notFullCond  *sync.Cond
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
		capacity:      options.capacity,
		lock:          sync.Mutex{},
	}

	queue.notEmptyCond = sync.NewCond(&queue.lock)
	queue.notFullCond = sync.NewCond(&queue.lock)

	if queue.capacity != nil {
		if len(queue.elements) > *queue.capacity {
			queue.elements = queue.elements[:*queue.capacity]
		}

		lenElements := len(elems)
		queue.initialLen = &lenElements
	}

	return queue
}

// ==================================Insertion=================================

// Put inserts the element to the tail the queue,
// while also increasing the queue size.
func (q *Blocking[T]) Put(elem T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.isFull() {
		fmt.Print("brad")
		q.notFullCond.Wait()
	}

	q.elements = append(q.elements, elem)

	q.notEmptyCond.Signal()
}

// Reset sets the queue elements index to 0. The queue will be in its initial
// state.
func (q *Blocking[T]) Reset() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.elementsIndex = 0

	if q.initialLen != nil {
		q.elements = q.elements[:*q.initialLen]
	}

	q.notEmptyCond.Broadcast()
}

// ===================================Removal==================================

// Take removes and returns the head of the elements queue.
// If no element is available it waits until the queue
// has an element available.
//
// It does not actually remove elements from the elements slice, but
// it's incrementing the underlying index.
func (q *Blocking[T]) Take() (v T) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.notFullCond.Signal()

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
	defer q.notFullCond.Signal()

	if q.isEmpty() {
		return v, ErrNoElementsAvailable
	}

	elem := q.elements[q.elementsIndex]

	q.elementsIndex++

	return elem, nil
}

// =================================Examination================================

// Peek retrieves but does not return the head of the queue.
func (q *Blocking[T]) Peek() T {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.isEmpty() {
		q.notEmptyCond.Wait()
	}

	elem := q.elements[q.elementsIndex]

	// send the not empty signal again in case any remove method waits.
	q.notEmptyCond.Signal()

	return elem
}

// Size returns the number of elements in the queue.
func (q *Blocking[T]) Size() int {
	q.lock.Lock()
	defer q.lock.Unlock()

	return len(q.elements) - q.elementsIndex
}

// ===================================Helpers==================================

func (q *Blocking[T]) getNextIndexOrWait() int {
	if !q.isEmpty() {
		return q.elementsIndex
	}

	q.notEmptyCond.Wait()

	return q.getNextIndexOrWait()
}

// ===========================To be used with mutexes==========================

func (q *Blocking[T]) isEmpty() bool {
	return q.elementsIndex >= len(q.elements)
}

func (q *Blocking[T]) isFull() bool {
	if q.capacity == nil {
		return false
	}

	return len(q.elements)-q.elementsIndex >= *q.capacity
}
