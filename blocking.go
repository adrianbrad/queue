package queue

import (
	"context"
	"sync"
)

// Blocking provides a read-only queue for a list of T.
//
// It supports operations for retrieving and adding elements to a FIFO queue.
// If there are no elements available the retrieve operations wait until
// elements are added to the queue.
type Blocking[T any] struct {
	// elements queue
	elements []T

	elementsChan  chan T
	refillMutex   sync.Mutex
	elementsIndex int
}

// NewBlocking returns an initialized Blocking Queue.
func NewBlocking[T any](elements []T) *Blocking[T] {
	c := make(chan T, len(elements))

	// load the elements into the buffered channel.
	for i := range elements {
		c <- elements[i]
	}

	return &Blocking[T]{
		elements:      elements,
		elementsChan:  c,
		refillMutex:   sync.Mutex{},
		elementsIndex: 0,
	}
}

// Take removes and returns the head of the elements queue.
// If no element is available it waits until the queue
//
// It does not actually remove elements from the elements slice, but
// it's incrementing the underlying index.
func (q *Blocking[T]) Take(
	ctx context.Context,
) (v T) {
	select {
	case v = <-q.elementsChan: // load the next element into the v variable
	case <-ctx.Done():
	}

	// return v which is either the default value for T or the next
	// element from the queue.
	return v
}

// Refill attempts to refill the queue with the elements added at
// initialization.
// If there is no room for new elements in the channel the method blocks
// until there is an available spot for the element or the context is closed.
//
// ! There is a chance that this method can block indefinitely if other
// threads are constantly reading from the queue, so a timeout context
// would be recommended.
func (q *Blocking[T]) Refill(ctx context.Context) {
	q.refillMutex.Lock()
	defer q.refillMutex.Unlock()

	// execute the loop until the elements channel is full.
	for i := q.elementsIndex; len(q.elementsChan) <= cap(q.elementsChan); i++ {
		// if the elements slice is consumed, reset the index and consume
		// it again from the start.
		if i == len(q.elements) {
			i = 0
		}

		select {
		// first of all check if the context was cancelled.
		case <-ctx.Done():
			return

		default:
			select {
			case q.elementsChan <- q.elements[i]:
				// successfully sent an element to the elements channel.

			case <-ctx.Done():
				return

			default:
				// channel is full, store the elements index and return.
				q.elementsIndex = i
				return
			}
		}
	}
}
