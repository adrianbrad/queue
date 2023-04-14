package queue

import (
	"sync"
)

// Ensure Priority implements the Queue interface.
var _ Queue[any] = (*Circular[any])(nil)

// Circular is a Queue implementation.
// A circular queue is a queue that uses a fixed-size slice as if it were connected end-to-end.
// When the queue is full, adding a new element to the queue overwrites the oldest element.
//
// Example:
// We have the following queue with a capacity of 3 elements: [1, 2, 3].
// If the tail of the queue is set to 0, as if we just added the element `3`,
// then the next element to be added to the queue will overwrite the element at index 0.
// So, if we add the element `4`, the queue will look like this: [4, 2, 3].
// If the head of the queue is set to 0, as if we never removed an element yet,
// then the next element to be removed from the queue will be the element at index 0, which is `4`.
type Circular[T comparable] struct {
	initialElements []T
	elems           []T
	head            int
	tail            int
	size            int

	// synchronization
	lock sync.RWMutex
}

// NewCircular creates a new Circular Queue containing the given elements.
func NewCircular[T comparable](
	givenElems []T,
	capacity int,
	opts ...Option,
) *Circular[T] {
	options := options{
		capacity: &capacity,
	}

	for _, o := range opts {
		o.apply(&options)
	}

	elems := make([]T, *options.capacity)

	copy(elems, givenElems)

	initialElems := make([]T, len(givenElems))

	copy(initialElems, givenElems)

	tail := 0

	size := len(elems)

	if len(initialElems) < len(elems) {
		tail = len(initialElems)
		size = len(initialElems)
	}

	return &Circular[T]{
		initialElements: initialElems,
		elems:           elems,
		head:            0,
		tail:            tail,
		size:            size,
		lock:            sync.RWMutex{},
	}
}

// ==================================Insertion=================================

// Offer adds an element into the queue.
// If the queue is full then the oldest item is overwritten.
func (q *Circular[T]) Offer(item T) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.size < len(q.elems) {
		q.size++
	}

	q.elems[q.tail] = item
	q.tail = (q.tail + 1) % len(q.elems)

	return nil
}

// Reset resets the queue to its initial state.
func (q *Circular[T]) Reset() {
	q.lock.Lock()
	defer q.lock.Unlock()

	copy(q.elems, q.initialElements)

	q.head = 0
	q.tail = 0
	q.size = len(q.initialElements)

	if len(q.initialElements) < len(q.elems) {
		q.tail = len(q.initialElements)
	}
}

// ===================================Removal==================================

// Get returns the element at the head of the queue.
func (q *Circular[T]) Get() (v T, _ error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.get()
}

// Clear removes all elements from the queue.
func (q *Circular[T]) Clear() []T {
	q.lock.Lock()
	defer q.lock.Unlock()

	elems := make([]T, 0, q.size)

	for {
		elem, err := q.get()
		if err != nil {
			break
		}

		elems = append(elems, elem)
	}

	// clear the queue
	q.head = 0
	q.tail = 0

	return elems
}

// Iterator returns an iterator over the elements in the queue.
// It removes the elements from the queue.
func (q *Circular[T]) Iterator() <-chan T {
	q.lock.RLock()
	defer q.lock.RUnlock()

	// use a buffered channel to avoid blocking the iterator.
	iteratorCh := make(chan T, q.size)

	// close the channel when the function returns.
	defer close(iteratorCh)

	// iterate over the elements and send them to the channel.
	for {
		elem, err := q.get()
		if err != nil {
			break
		}

		iteratorCh <- elem
	}

	return iteratorCh
}

// =================================Examination================================

// IsEmpty returns true if the queue is empty.
func (q *Circular[T]) IsEmpty() bool {
	q.lock.RLock()
	defer q.lock.RUnlock()

	return q.isEmpty()
}

// Contains returns true if the queue contains the given element.
func (q *Circular[T]) Contains(elem T) bool {
	q.lock.RLock()
	defer q.lock.RUnlock()

	if q.isEmpty() {
		return false // queue is empty, item not found
	}

	for i := q.head; i < q.size; i++ {
		idx := (q.head + i) % len(q.elems)

		if q.elems[idx] == elem {
			return true // item found
		}
	}

	return false // item not found
}

// Peek returns the element at the head of the queue.
func (q *Circular[T]) Peek() (v T, _ error) {
	q.lock.RLock()
	defer q.lock.RUnlock()

	if q.isEmpty() {
		return v, ErrNoElementsAvailable
	}

	return q.elems[q.head], nil
}

// Size returns the number of elements in the queue.
func (q *Circular[T]) Size() int {
	q.lock.RLock()
	defer q.lock.RUnlock()

	return q.size
}

// ===================================Helpers==================================

// get returns the element at the head of the queue.
func (q *Circular[T]) get() (v T, _ error) {
	if q.isEmpty() {
		return v, ErrNoElementsAvailable
	}

	item := q.elems[q.head]
	q.head = (q.head + 1) % len(q.elems)
	q.size--

	return item, nil
}

// isEmpty returns true if the queue is empty.
func (q *Circular[T]) isEmpty() bool {
	return q.size == 0
}
