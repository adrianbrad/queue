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

Queue is a Go package that provides multiple thread-safe generic queue implementations.

A queue is a sequence of entities that is open at both ends where he elements are 
added(enqueued) to the tail(back) of the queue and removed(dequeued) from the head(front) of the queue.

### Notable characteristics
- Usable zero values.
- The queues can be reset to their initial state.
- Implementations satisfy the `Queue` interface.
- Example tests are provided.

## Blocking Queue
- FIFO Ordering 
- Waits for the queue have elements available before retrieving from it.
- Implemented using `sync.Cond` from standard library.

## Priority Queue
- Order based on the `Less` method implemented by elements.
- Waits for the queue have elements available before retrieving from it.
- Implemented using `container/heap` standard library package.