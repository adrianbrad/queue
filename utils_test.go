package queue_test

import (
	"github.com/adrianbrad/queue"
)

func drainQueue[T any](q queue.Queue[T]) []T {
	size := q.Size()

	elems := make([]T, size)

	var err error

	for i := 0; i < size; i++ {
		elems[i], err = q.Get()
		if err != nil {
			return nil
		}
	}

	return elems
}
