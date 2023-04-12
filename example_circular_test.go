package queue_test

import (
	"fmt"

	"github.com/adrianbrad/queue"
)

func ExampleCircular() {
	elems := []int{1, 2, 3}

	const capacity = 4

	priorityQueue := queue.NewCircular(
		elems,
		capacity,
	)

	containsTwo := priorityQueue.Contains(2)
	fmt.Println("Contains 2:", containsTwo)

	size := priorityQueue.Size()
	fmt.Println("Size:", size)

	if err := priorityQueue.Offer(4); err != nil {
		fmt.Println("Offer err: ", err)
		return
	}

	nextElem, err := priorityQueue.Peek()
	if err != nil {
		fmt.Println("Peek err: ", err)
		return
	}

	fmt.Println("Peek:", nextElem)

	if err := priorityQueue.Offer(5); err != nil {
		fmt.Println("Offer err: ", err)
		return
	}

	fmt.Println("Offered 5")

	if err := priorityQueue.Offer(6); err != nil {
		fmt.Println("Offer err: ", err)
		return
	}

	fmt.Println("Offered 6")

	clearElems := priorityQueue.Clear()
	fmt.Println("Clear:", clearElems)

	fmt.Println("Offered 7")

	if err := priorityQueue.Offer(7); err != nil {
		fmt.Println("Offer err: ", err)
		return
	}

	elem, err := priorityQueue.Get()
	if err != nil {
		fmt.Println("Get err: ", err)
		return
	}

	fmt.Println("Get:", elem)

	// Output:
	// Contains 2: true
	// Size: 3
	// Peek: 1
	// Offered 5
	// Offered 6
	// Clear: [5 6 3 4]
	// Offered 7
	// Get: 7
}
