package queue_test

import (
	"fmt"

	"github.com/adrianbrad/queue"
)

var _ queue.Lesser = (*intValAscending)(nil)

type intValAscending int

func (i intValAscending) Less(other any) bool {
	return i < other.(intValAscending)
}

var _ queue.Lesser = (*intValDescending)(nil)

type intValDescending int

func (i intValDescending) Less(other any) bool {
	return i > other.(intValDescending)
}

func ExamplePriority() {
	fmt.Println("Ascending:")

	elemsAsc := []intValAscending{2, 4, 1}

	pAsc := queue.NewPriority(elemsAsc, queue.WithCapacity(4))

	if err := pAsc.Offer(3); err != nil {
		fmt.Printf("offer err: %s\n", err)
		return
	}

	fmt.Println(pAsc.Offer(5))
	fmt.Println(drainQueue[intValAscending](pAsc))
	fmt.Println(pAsc.Get())

	fmt.Printf("\nDescending:\n")

	elemsDesc := []intValDescending{2, 4, 1}

	pDesc := queue.NewPriority(elemsDesc, queue.WithCapacity(4))

	if err := pDesc.Offer(3); err != nil {
		fmt.Printf("offer err: %s\n", err)
		return
	}

	fmt.Println(drainQueue[intValDescending](pDesc))

	// Output:
	// Ascending:
	// queue is full
	// [1 2 3 4]
	// 0 no elements available in the queue
	//
	// Descending:
	// [4 3 2 1]
}

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
