package queue_test

import (
	"fmt"

	"github.com/adrianbrad/queue"
)

// runExampleDemo exercises a queue through a shared sequence of operations
// used by ExampleLinked and ExamplePriority. The logical output differs per
// implementation (e.g. clear order for priority vs. FIFO), so each Example
// keeps its own // Output: block.
func runExampleDemo(q queue.Queue[int]) {
	fmt.Println("Contains 2:", q.Contains(2))
	fmt.Println("Size:", q.Size())

	if err := q.Offer(3); err != nil {
		fmt.Println("Offer err: ", err)
		return
	}

	fmt.Println("Empty before clear:", q.IsEmpty())
	fmt.Println("Clear:", q.Clear())
	fmt.Println("Empty after clear:", q.IsEmpty())

	if err := q.Offer(5); err != nil {
		fmt.Println("Offer err: ", err)
		return
	}

	elem, err := q.Get()
	if err != nil {
		fmt.Println("Get err: ", err)
		return
	}

	fmt.Println("Get:", elem)
}
