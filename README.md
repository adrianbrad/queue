# queue ![GitHub release](https://img.shields.io/github/v/tag/adrianbrad/queue)

[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/adrianbrad/queue)](https://github.com/adrianbrad/queue)
[![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/adrianbrad/queue)

[![CodeFactor](https://www.codefactor.io/repository/github/adrianbrad/queue/badge)](https://www.codefactor.io/repository/github/adrianbrad/queue)
[![Go Report Card](https://goreportcard.com/badge/github.com/adrianbrad/queue)](https://goreportcard.com/report/github.com/adrianbrad/queue)
[![codecov](https://codecov.io/gh/adrianbrad/queue/branch/main/graph/badge.svg)](https://codecov.io/gh/adrianbrad/queue)

[![lint-test](https://github.com/adrianbrad/queue/actions/workflows/lint-test.yaml/badge.svg)](https://github.com/adrianbrad/queue/actions?query=workflow%3Alint-test)
[![grype](https://github.com/adrianbrad/queue/actions/workflows/grype.yaml/badge.svg)](https://github.com/adrianbrad/queue/actions?query=workflow%3Agrype)
[![codeql](https://github.com/adrianbrad/queue/actions/workflows/codeql.yaml/badge.svg)](https://github.com/adrianbrad/queue/actions?query=workflow%3ACodeQL)

---

### Overview 

The queue package provides multiple thread-safe generic queue implementations in Go.

A queue is a sequence of entities that is open at both ends where the elements are
added (enqueued) to the tail (back) of the queue and removed (dequeued) from the head (front) of the queue.

Queues implemented in this package are designed to be easy to use and provide a consistent API.

Benchmarks and Example tests can be found in this package.

### Features
The queue package provides two types of queues:

- Blocking Queue: FIFO Ordering, provides blocking and non-blocking methods. The non-blocking methods return an error. Implemented using sync.Cond from the standard library.

- Priority Queue: Order based on the less function provided at construction. Implemented using container/heap standard library package.

### Installation
To add this package as a dependency to your project, run:

```
go get -u github.com/adrianbrad/queue
```

### Import
To use this package in your project, you can import it as follows:

```go
import "github.com/adrianbrad/queue"
```

### Usage


#### Queue Interface

```go
// Queue is a generic queue interface, defining the methods that all queues must implement.
type Queue[T comparable] interface {
	Get() (T, error)
	Offer(T) error
	Reset()
	Peek() (T, error)
	Size() int
	IsEmpty() bool
	Iterator() <-chan T
	Clear() []T
}
```

#### Quick Start

##### Blocking Queue

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

	fmt.Println("elem: ", elem) // elem: 2
}
```

##### Priority Queue

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

## Benchmarks 

Results as of 3rd of February 2023.

```
BenchmarkBlockingQueue/Peek-12          63275360                19.44 ns/op            0 B/op          0 allocs/op
BenchmarkBlockingQueue/Get_Offer-12     19066974                69.67 ns/op           40 B/op          0 allocs/op
BenchmarkBlockingQueue/Offer-12         36569245                37.86 ns/op           41 B/op          0 allocs/op
BenchmarkPriorityQueue/Peek-12          66765319                15.86 ns/op            0 B/op          0 allocs/op
BenchmarkPriorityQueue/Get_Offer-12     16677442                71.33 ns/op            0 B/op          0 allocs/op
BenchmarkPriorityQueue/Offer-12         20044909                58.58 ns/op           55 B/op          0 allocs/op
```