package queue_test

import (
	"fmt"

	"github.com/adrianbrad/queue"
)

func ExampleLinked() {
	elems := []int{2, 4, 1}

	priorityQueue := queue.NewLinked(
		elems,
	)

	containsTwo := priorityQueue.Contains(2)
	fmt.Println("Contains 2:", containsTwo)

	size := priorityQueue.Size()
	fmt.Println("Size:", size)

	if err := priorityQueue.Offer(3); err != nil {
		fmt.Println("Offer err: ", err)
		return
	}

	empty := priorityQueue.IsEmpty()
	fmt.Println("Empty before clear:", empty)

	clearElems := priorityQueue.Clear()
	fmt.Println("Clear:", clearElems)

	empty = priorityQueue.IsEmpty()
	fmt.Println("Empty after clear:", empty)

	if err := priorityQueue.Offer(5); err != nil {
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
	// Empty before clear: false
	// Clear: [2 4 1 3]
	// Empty after clear: true
	// Get: 5
}
