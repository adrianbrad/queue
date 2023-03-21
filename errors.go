package queue

import (
	"errors"
)

var (
	// ErrNoElementsAvailable is an error returned whenever there are no
	// elements available to be extracted from a queue.
	ErrNoElementsAvailable = errors.New("no elements available in the queue")

	// ErrQueueIsFull is an error returned whenever the queue is full and there
	// is an attempt to add an element to it.
	ErrQueueIsFull = errors.New("queue is full")
)
