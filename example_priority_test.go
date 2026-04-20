package queue_test

import (
	"github.com/adrianbrad/queue"
)

func ExamplePriority() {
	elems := []int{2, 4, 1}

	priorityQueue := queue.NewPriority(
		elems,
		func(elem, otherElem int) bool {
			return elem < otherElem
		},
		queue.WithCapacity(4),
	)

	runExampleDemo(priorityQueue)

	// Output:
	// Contains 2: true
	// Size: 3
	// Empty before clear: false
	// Clear: [1 2 3 4]
	// Empty after clear: true
	// Get: 5
}
