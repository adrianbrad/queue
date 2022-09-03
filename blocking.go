package queue

import (
	"context"

	"go.uber.org/atomic"
)

// Blocking provides a read-only queue for a list of T.
//
// It provides a Take method for popping elements from the queue head (FIFO).
// If there are no elements available the Take method blocks until Reset
// is called.
//
// Reset refills the elements queue.
type Blocking[T any] struct {
	// elements queue
	elements []T

	// sync stores the broadcast channel and the index of the last element.
	sync atomic.Pointer[sync]
}

type sync struct {
	// broadcastChannel stores a channel which, when closed, emits a signal
	// to all goroutines listening to it in Blocking.Take.
	broadcastChannel chan struct{} // index of the last element

	// atomic.Uintptr instead of Uint32 or Uint64, as it correlates with the
	// max length for slices which is
	// defined by int (int32, max = 1<<31 - 1 on 32bit platforms and
	// int64, max = 1<<63 - 1 on 64bit platforms)
	index atomic.Uintptr
}

var zeroUintptr = atomic.Uintptr{}

// NewBlocking returns an initialized Blocking Queue.
func NewBlocking[T any](elements []T) *Blocking[T] {
	return &Blocking[T]{
		elements: elements,
		sync: *atomic.NewPointer(&sync{
			broadcastChannel: make(chan struct{}),
			index:            zeroUintptr,
		}),
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
	s := q.sync.Load()

	newIndex := s.index.Inc()

	// check if there is an element available.
	if int(newIndex) > len(q.elements) {
		// if no elements are available wait for Reset or context close.
		select {
		// wait for the reset signal.
		// acts like sync.Cond.Wait but with a channel.
		case <-s.broadcastChannel: // q.index is 0 here
			return q.Take(ctx)

		// caller context is canceled, return default value for T and no err.
		case <-ctx.Done():
			return v
		}
	}

	// if there is an element available, return it.
	return q.elements[newIndex-1]
}

// Peek returns but does not remove the element at the head of the queue.
func (q *Blocking[T]) Peek() T {
	// this can produce inconsistencies with Take(), as the index
	// could be read before or after the increment at line :52.
	return q.elements[q.sync.Load().index.Load()]
}

// Reset notifies every blocking Take routine that index can be reset.
// nolint: revive // line too long
// inspiration from pre go 1.18(generics) code: https://gist.github.com/zviadm/c234426882bfc8acba88f3503edaaa36#file-cond2-go-L54
func (q *Blocking[_]) Reset() {
	// replace the sync object, with a new one
	// containing a fresh channel and index.
	// save the old sync object in order to close the broadcast channel and
	// resume execution of all goroutines waiting in the select block
	// in the Take method.
	oldSync := q.sync.Swap(&sync{
		broadcastChannel: make(chan struct{}),
		index:            zeroUintptr,
	})

	// close the old broadcast channel thus resuming all the sleeping
	// goroutines waiting in the first select case of the Take method.
	//
	// this acts like a sync.Cond.Broadcast().
	close(oldSync.broadcastChannel)
}
