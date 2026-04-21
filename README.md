# queue

![GitHub release](https://img.shields.io/github/v/tag/adrianbrad/queue)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/adrianbrad/queue)](https://github.com/adrianbrad/queue)
[![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/adrianbrad/queue)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

[![CodeFactor](https://www.codefactor.io/repository/github/adrianbrad/queue/badge)](https://www.codefactor.io/repository/github/adrianbrad/queue)
[![Go Report Card](https://goreportcard.com/badge/github.com/adrianbrad/queue)](https://goreportcard.com/report/github.com/adrianbrad/queue)
[![codecov](https://codecov.io/gh/adrianbrad/queue/branch/main/graph/badge.svg)](https://codecov.io/gh/adrianbrad/queue)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/adrianbrad/queue/badge)](https://scorecard.dev/viewer/?uri=github.com/adrianbrad/queue)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/12607/badge)](https://www.bestpractices.dev/projects/12607)

[![lint-test](https://github.com/adrianbrad/queue/actions/workflows/lint-test.yaml/badge.svg)](https://github.com/adrianbrad/queue/actions?query=workflow%3Alint-test)
[![grype](https://github.com/adrianbrad/queue/actions/workflows/grype.yaml/badge.svg)](https://github.com/adrianbrad/queue/actions?query=workflow%3Agrype)
[![codeql](https://github.com/adrianbrad/queue/actions/workflows/codeql.yaml/badge.svg)](https://github.com/adrianbrad/queue/actions?query=workflow%3ACodeQL)

---

Thread-safe, generic FIFO, priority, circular, linked, and delay queues for Go.

## Features

- Five queue flavours behind one `Queue[T comparable]` interface, so you can swap implementations without changing call sites.
- Generic types with no reflection; zero third-party dependencies.
- Steady-state zero-alloc reads on every queue and zero-alloc offer/get on `Circular`, `Linked`, `Priority`, and `Delay`.
- Blocking variants (`OfferWait`, `GetWait`, `PeekWait`) for producer/consumer workloads.
- `Delay` queue for timers, retry scheduling, and TTL expiry.
- 100% test coverage and race-tested in CI.

Full API reference at **[pkg.go.dev/github.com/adrianbrad/queue](https://pkg.go.dev/github.com/adrianbrad/queue)**.

<!-- TOC -->
* [queue](#queue)
  * [Features](#features)
  * [Installation](#installation)
  * [Import](#import)
  * [Quick start](#quick-start)
  * [Choosing a queue](#choosing-a-queue)
  * [Usage](#usage)
    * [Queue Interface](#queue-interface)
    * [Blocking Queue](#blocking-queue)
    * [Priority Queue](#priority-queue)
    * [Circular Queue](#circular-queue)
    * [Linked Queue](#linked-queue)
    * [Delay Queue](#delay-queue)
  * [Benchmarks](#benchmarks)
  * [Contributing](#contributing)
  * [Security](#security)
  * [License](#license)
<!-- TOC -->

## Installation
To add this package as a dependency to your project, run:

```shell
go get -u github.com/adrianbrad/queue
```

## Import
To use this package in your project, you can import it as follows:

```go
import "github.com/adrianbrad/queue"
```

## Quick start

```go
package main

import (
	"fmt"

	"github.com/adrianbrad/queue"
)

func main() {
	q := queue.NewLinked([]int{1, 2, 3})

	_ = q.Offer(4)

	// prints 1, then 2, then 3, then 4 (one per iteration)
	for !q.IsEmpty() {
		v, _ := q.Get()
		fmt.Println(v)
	}
}
```

Every implementation satisfies the same `Queue[T comparable]` interface. Pick a different constructor (`NewBlocking`, `NewPriority`, `NewCircular`, `NewDelay`) without changing the call sites.

## Choosing a queue

| Queue      | Ordering            | Capacity                                      | Blocks?                                            | Pick this when…                                                                                 |
|------------|---------------------|-----------------------------------------------|----------------------------------------------------|-------------------------------------------------------------------------------------------------|
| `Blocking` | FIFO                | Optional; `Offer` errors on full              | Yes, via `OfferWait`, `GetWait`, `PeekWait`        | You want a classic producer-consumer queue with backpressure and blocking semantics.            |
| `Priority` | Custom (less func)  | Optional; `Offer` errors on full              | No                                                 | Order depends on a computed value (smallest deadline, highest score, lexicographic, etc).       |
| `Circular` | FIFO                | Required; `Offer` **overwrites the oldest**   | No                                                 | You want fixed memory and the most recent N items; dropping older entries is acceptable.        |
| `Linked`   | FIFO                | None (unbounded)                              | No                                                 | You need an unbounded FIFO and don't want to pick a capacity up front.                          |
| `Delay`    | By deadline         | Optional; `Offer` errors on full              | `GetWait` sleeps until the head's deadline passes  | Items should become available at a future time (timers, retry scheduling, TTL expiry).          |

## Usage

### Queue Interface

```go
// Queue is a generic queue interface, defining the methods that all queues must implement.
type Queue[T comparable] interface {
	// Get retrieves and removes the head of the queue.
	Get() (T, error)

	// Offer inserts the element to the tail of the queue.
	Offer(T) error

	// Reset sets the queue to its initial state.
	Reset()

	// Contains returns true if the queue contains the element.
	Contains(T) bool

	// Peek retrieves but does not remove the head of the queue.
	Peek() (T, error)

	// Size returns the number of elements in the queue.
	Size() int

	// IsEmpty returns true if the queue is empty.
	IsEmpty() bool

	// Iterator returns a channel that will be filled with the elements
	Iterator() <-chan T

	// Clear removes all elements from the queue.
	Clear() []T
}
```

### Blocking Queue

Blocking queue is a FIFO ordered data structure. Both blocking and non-blocking methods are implemented.
Blocking methods wait for the queue to have available items when dequeuing, and wait for a slot to become available in case the queue is full when enqueuing.
The non-blocking methods return an error if an element cannot be added or removed.
Implemented using sync.Cond from the standard library.

```go
package main

import (
	"fmt"

	"github.com/adrianbrad/queue"
)

func main() {
	elems := []int{2, 3}

	blockingQueue := queue.NewBlocking(elems, queue.WithCapacity(3))

	containsTwo := blockingQueue.Contains(2)
	fmt.Println(containsTwo) // true

	size := blockingQueue.Size()
	fmt.Println(size) // 2

	empty := blockingQueue.IsEmpty()
	fmt.Println(empty) // false

	if err := blockingQueue.Offer(1); err != nil {
		// handle err
	}

	elem, err := blockingQueue.Get()
	if err != nil {
		// handle err
	}

	fmt.Printf("elem: %d\n", elem) // elem: 2
}
```

### Priority Queue

Priority Queue is a data structure where the order of the elements is given by a less function provided at construction.
Implemented over an internal min-heap.

```go
package main

import (
	"fmt"

	"github.com/adrianbrad/queue"
)

func main() {
	elems := []int{2, 3, 4}

	priorityQueue := queue.NewPriority(
		elems,
		func(elem, otherElem int) bool { return elem < otherElem },
	)

	containsTwo := priorityQueue.Contains(2)
	fmt.Println(containsTwo) // true

	size := priorityQueue.Size()
	fmt.Println(size) // 3

	empty := priorityQueue.IsEmpty()
	fmt.Println(empty) // false

	if err := priorityQueue.Offer(1); err != nil {
		// handle err
	}

	elem, err := priorityQueue.Get()
	if err != nil {
		// handle err
	}

	fmt.Printf("elem: %d\n", elem) // elem: 1
}
```

### Circular Queue

Circular Queue is a fixed size FIFO ordered data structure. When the queue is full, adding a new element to the queue overwrites the oldest element.

Example:
We have the following queue with a capacity of 3 elements: [1, 2, 3].
If the tail of the queue is set to 0, as if we just added the element `3`,
the next element to be added to the queue will overwrite the element at index 0.
So, if we add the element `4`, the queue will look like this: [4, 2, 3].
If the head of the queue is set to 0, as if we never removed an element yet,
then the next element to be removed from the queue will be the element at index 0, which is `4`.

```go
package main

import (
	"fmt"

	"github.com/adrianbrad/queue"
)

func main() {
	elems := []int{2, 3, 4}

	circularQueue := queue.NewCircular(elems, 3)

	containsTwo := circularQueue.Contains(2)
	fmt.Println(containsTwo) // true

	size := circularQueue.Size()
	fmt.Println(size) // 3

	empty := circularQueue.IsEmpty()
	fmt.Println(empty) // false

	if err := circularQueue.Offer(1); err != nil {
		// handle err
	}

	elem, err := circularQueue.Get()
	if err != nil {
		// handle err
	}

	fmt.Printf("elem: %d\n", elem) // elem: 1
}
```

### Linked Queue

A linked queue, implemented as a singly linked list, offering O(1)
time complexity for enqueue and dequeue operations. The queue maintains pointers
to both the head (front) and tail (end) of the list for efficient operations
without the need for traversal.

```go
package main

import (
	"fmt"

	"github.com/adrianbrad/queue"
)

func main() {
	elems := []int{2, 3, 4}

	linkedQueue := queue.NewLinked(elems)

	containsTwo := linkedQueue.Contains(2)
	fmt.Println(containsTwo) // true

	size := linkedQueue.Size()
	fmt.Println(size) // 3

	empty := linkedQueue.IsEmpty()
	fmt.Println(empty) // false

	if err := linkedQueue.Offer(1); err != nil {
		// handle err
	}

	elem, err := linkedQueue.Get()
	if err != nil {
		// handle err
	}

	fmt.Printf("elem: %d\n", elem) // elem: 2
}
```

### Delay Queue

A `Delay` queue is a priority queue where each element becomes dequeuable at a deadline computed by a caller-supplied function at `Offer` time. `Get` returns `ErrNoElementsAvailable` until the head's deadline has passed; `GetWait` sleeps until it does. Useful for timers, retry scheduling, and TTL expiry.

```go
package main

import (
	"fmt"
	"time"

	"github.com/adrianbrad/queue"
)

type task struct {
	id    int
	runAt time.Time
}

func main() {
	now := time.Now()

	delayQueue := queue.NewDelay(
		[]task{
			{id: 1, runAt: now.Add(20 * time.Millisecond)},
			{id: 2, runAt: now.Add(5 * time.Millisecond)},
		},
		func(t task) time.Time { return t.runAt },
	)

	size := delayQueue.Size()
	fmt.Println(size) // 2

	// Non-blocking: not due yet.
	if _, err := delayQueue.Get(); err != nil {
		// err == queue.ErrNoElementsAvailable
	}

	// Blocking: returns as soon as the head's deadline passes.
	next := delayQueue.GetWait()
	fmt.Printf("next: %d\n", next.id) // next: 2
}
```

## Benchmarks

Run locally with `go test -bench=. -benchmem -benchtime=3s -count=3`. Reported numbers are per-operation timings and allocations; absolute values vary by hardware, but the shape (zero-alloc reads everywhere, zero-alloc offer/get for Circular, Linked, Priority, and Delay) should be stable.

```text
BenchmarkBlockingQueue/Peek                  3.8 ns/op       0 B/op   0 allocs/op
BenchmarkBlockingQueue/Get_Offer            22.9 ns/op       8 B/op   1 allocs/op
BenchmarkBlockingQueue/Offer                13.0 ns/op      49 B/op   0 allocs/op
BenchmarkCircularQueue/Peek                  3.9 ns/op       0 B/op   0 allocs/op
BenchmarkCircularQueue/Get_Offer            13.9 ns/op       0 B/op   0 allocs/op
BenchmarkCircularQueue/Offer                 6.5 ns/op       0 B/op   0 allocs/op
BenchmarkLinkedQueue/Peek                    3.9 ns/op       0 B/op   0 allocs/op
BenchmarkLinkedQueue/Get_Offer              14.7 ns/op       0 B/op   0 allocs/op
BenchmarkLinkedQueue/Offer                  22.7 ns/op      16 B/op   1 allocs/op
BenchmarkPriorityQueue/Peek                  3.9 ns/op       0 B/op   0 allocs/op
BenchmarkPriorityQueue/Get_Offer            18.1 ns/op       0 B/op   0 allocs/op
BenchmarkPriorityQueue/Offer                17.1 ns/op      48 B/op   0 allocs/op
BenchmarkDelayQueue/Peek                     4.1 ns/op       0 B/op   0 allocs/op
BenchmarkDelayQueue/Get_Offer               52.4 ns/op       0 B/op   0 allocs/op
BenchmarkDelayQueue/Offer                   63.5 ns/op     315 B/op   0 allocs/op
```

## Contributing

PRs welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for the workflow, coding conventions, and the 100% coverage requirement. Ask questions by opening a GitHub issue.

## Security

Please report security issues privately following [SECURITY.md](SECURITY.md) rather than opening a public issue.

## License

MIT. See [LICENSE](LICENSE). Maintained by [@adrianbrad](https://github.com/adrianbrad).
