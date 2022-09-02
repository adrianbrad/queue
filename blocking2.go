package queue

import (
	"context"
	"go.uber.org/atomic"
)

type Blocking2[T any] struct {
	elements []T

	c *atomic.Pointer[chan T]
}

func NewBlocking2[T any](elements []T) *Blocking2[T] {
	c := make(chan T, len(elements))
	for i := range elements {
		c <- elements[i]
	}

	return &Blocking2[T]{
		elements: elements,
		c:        atomic.NewPointer(&c),
	}
}

func (q *Blocking2[T]) Take(
	ctx context.Context,
) (v T) {
	select {
	case e, ok := <-*q.c.Load():
		if !ok {
			return q.Take(ctx)
		}

		v = e
	case <-ctx.Done():
	}

	return v
}

func (q *Blocking2[T]) Reset() {
	newC := make(chan T, len(q.elements))

	for i := range q.elements {
		newC <- q.elements[i]
	}

	oldC := q.c.Swap(&newC)

	close(*oldC)
}
