package queue

import (
	"context"

	"go.uber.org/atomic"
)

// Blocking provides a read-only storage for a list of T.
// It provides a Get method for reading elements one by one, in FIFO order.
// If there are no elements available the Get method blocks until Reset
// is called.
// Reset refills the elements queue.
type Blocking[T any] struct {
	// elements queue
	elements []T

	// index of the last element
	index *atomic.Uint32

	// broadcastChannelPtr stores a pointer to a channel which serves as
	// a broadcast channel.
	broadcastChannelPtr *atomic.Pointer[chan struct{}]
}

// NewBlocking returns an initialized Blocking.
func NewBlocking[T any](elements []T) *Blocking[T] {
	broadcastChannel := make(chan struct{})

	return &Blocking[T]{
		elements: elements,
		index:    atomic.NewUint32(0),
		broadcastChannelPtr: atomic.
			NewPointer[chan struct{}](&broadcastChannel),
	}
}

// Get returns the next element.
// It acts as a pop from queue where the queue is the elements slice.
// * Pop from queue deletes and returns the leftmost element.
//
// It does not actually remove elements from the elements slice, pop
// is implemented with the help of an index.
func (s *Blocking[T]) Get(
	ctx context.Context,
) (v T, _ error) {
	newIndex := s.index.Inc()

	// check if we have available elements
	if int(newIndex) > len(s.elements) {
		// if no elements are available wait for Reset or context close.
		select {
		// wait for the reset signal.
		// acts like sync.Cond.Wait but with a channel.
		case <-*s.broadcastChannelPtr.Load(): // s.index is 0 here
			return s.Get(ctx)

		// caller context is canceled, return default value for T and no err.
		case <-ctx.Done():
			return v, nil
		}
	}

	return s.elements[newIndex-1], nil
}

// Reset notifies every blocking Get routine that index can be reset.
// nolint: revive // line too long
// inspiration from pre go 1.18(generics) code: https://gist.github.com/zviadm/c234426882bfc8acba88f3503edaaa36#file-cond2-go-L54
func (s *Blocking[_]) Reset() error {
	// create a new signal channel
	newBroadcastChannel := make(chan struct{})

	// place the new broadcast channel in place of the old signal channel,
	// retrieve the old broadcast channel in order to close it and continue
	// execution of all goroutines waiting for the select in the Get method.
	oldBroadcastChannel := s.broadcastChannelPtr.Swap(&newBroadcastChannel)

	// reset elements index.
	s.index.Store(0)

	// close the old broadcast channel thus starting all the sleeping
	// goroutines waiting in the first select case of the Get method.
	//
	// this acts like a sync.Cond.Broadcast().
	close(*oldBroadcastChannel)

	return nil
}
