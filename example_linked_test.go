package queue_test

import (
	"github.com/adrianbrad/queue"
)

func ExampleLinked() {
	elems := []int{2, 4, 1}

	linkedQueue := queue.NewLinked(elems)

	runExampleDemo(linkedQueue)

	// Output:
	// Contains 2: true
	// Size: 3
	// Empty before clear: false
	// Clear: [2 4 1 3]
	// Empty after clear: true
	// Get: 5
}
