package queue

import (
	"errors"
)

// ErrNoElementsAvailable is an error returned whenever there are no elements
// available to be extracted from a queue.
var ErrNoElementsAvailable = errors.New("no elements available in the queue")
