package queue_test

import (
	"fmt"

	"github.com/adrianbrad/queue"
)

func ExamplePriority() {
	fmt.Println("Ascending:")

	elemsAsc := []int{2, 4, 1}

	pAsc := queue.NewPriority(
		elemsAsc,
		func(elem, otherElem int) bool {
			return elem < otherElem
		},
		queue.WithCapacity(4),
	)

	if err := pAsc.Offer(3); err != nil {
		fmt.Printf("offer err: %s\n", err)
		return
	}

	fmt.Println(pAsc.Offer(5))
	fmt.Println(drainQueue[int](pAsc))
	fmt.Println(pAsc.Get())

	fmt.Printf("\nDescending:\n")

	elemsDesc := []int{2, 4, 1}

	pDesc := queue.NewPriority(
		elemsDesc,
		func(elem, otherElem int) bool {
			return elem > otherElem
		},
		queue.WithCapacity(4),
	)

	if err := pDesc.Offer(3); err != nil {
		fmt.Printf("offer err: %s\n", err)
		return
	}

	fmt.Println(drainQueue[int](pDesc))

	// Output:
	// Ascending:
	// queue is full
	// [1 2 3 4]
	// 0 no elements available in the queue
	//
	// Descending:
	// [4 3 2 1]
}
