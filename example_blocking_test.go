package queue_test

import (
	"fmt"
	"time"

	"github.com/adrianbrad/queue"
)

func ExampleBlocking() {
	elems := []int{1, 2, 3}

	blockingQueue := queue.NewBlocking(elems, queue.WithCapacity(4))

	fmt.Println(drainQueue[int](blockingQueue))

	var (
		elem  int
		after time.Duration
	)

	done := make(chan struct{})

	start := time.Now()

	// this function waits for a new element to be available in the queue.
	go func() {
		defer close(done)

		elem = blockingQueue.GetWait()
		after = time.Since(start)
	}()

	time.Sleep(time.Millisecond)

	// insert a new element into the queue.
	fmt.Println("Offer err:", blockingQueue.Offer(4))

	<-done

	fmt.Printf(
		"Elem %d received after %s",
		elem,
		after.Round(time.Millisecond),
	)

	// Output:
	// [1 2 3]
	// Offer err: <nil>
	// Elem 4 received after 1ms
}
