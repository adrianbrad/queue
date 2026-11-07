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

	head, err := blockingQueue.Peek()
	if err != nil {
		fmt.Println("Peek err:", err)
		return
	}

	fmt.Println("Peeked elem:", head)

	empty := blockingQueue.IsEmpty()
	fmt.Println("Empty before clear:", empty)

	clearElems := blockingQueue.Clear()
	fmt.Println("Clear:", clearElems)

	empty = blockingQueue.IsEmpty()
	fmt.Println("Empty after clear:", empty)

	// Get does not block: on an empty queue it returns an error.
	if _, err := blockingQueue.Get(); err != nil {
		fmt.Println("Get err:", err)
	}

	received := make(chan int)

	// GetWait blocks until an element is available, so it is called from
	// another goroutine; it returns only once Offer below has run.
	go func() {
		received <- blockingQueue.GetWait()
	}()

	if err := blockingQueue.Offer(4); err != nil {
		fmt.Println("Offer err:", err)
		return
	}

	fmt.Println("GetWait received:", <-received)

	// Output:
	// Contains 3: true
	// Size: 3
	// Peeked elem: 1
	// Empty before clear: false
	// Clear: [1 2 3]
	// Empty after clear: true
	// Get err: no elements available in the queue
	// GetWait received: 4
}
