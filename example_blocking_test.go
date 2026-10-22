package queue_test

import (
	"fmt"

	"github.com/adrianbrad/queue"
)

func ExampleBlocking() {
	elems := []int{1, 2, 3}

	blockingQueue := queue.NewBlocking(elems, queue.WithCapacity(4))

	containsThree := blockingQueue.Contains(3)
	fmt.Println("Contains 3:", containsThree)

	size := blockingQueue.Size()
	fmt.Println("Size:", size)

	empty := blockingQueue.IsEmpty()
	fmt.Println("Empty before clear:", empty)

	clearElems := blockingQueue.Clear()
	fmt.Println("Clear:", clearElems)

	empty = blockingQueue.IsEmpty()
	fmt.Println("Empty after clear:", empty)

	var elem int
	done := make(chan struct{})

	// This function waits for a new element to be available in the queue.
	go func() {
		defer close(done)
		elem = blockingQueue.GetWait()
	}()

	// insert a new element into the queue.
	if err := blockingQueue.Offer(4); err != nil {
		fmt.Println("Offer err:", err)
		return
	}

	nextElem, err := blockingQueue.Peek()
	if err != nil {
		fmt.Println("Peek err:", err)
		return
	}

	fmt.Println("Peeked elem:", nextElem)

	<-done

	fmt.Println("Elem received:", elem)

	// Output:
	// Contains 3: true
	// Size: 3
	// Empty before clear: false
	// Clear: [1 2 3]
	// Empty after clear: true
	// Peeked elem: 4
	// Elem received: 4
}
