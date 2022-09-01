# queue ![GitHub release](https://img.shields.io/github/v/release/adrianbrad/queue)

[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/adrianbrad/queue)](https://github.com/adrianbrad/queue)
[![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/adrianbrad/queue)

[![CodeFactor](https://www.codefactor.io/repository/github/adrianbrad/queue/badge)](https://www.codefactor.io/repository/github/adrianbrad/queue)
[![Go Report Card](https://goreportcard.com/badge/github.com/adrianbrad/queue)](https://goreportcard.com/report/github.com/adrianbrad/queue)
[![codecov](https://codecov.io/gh/adrianbrad/queue/branch/main/graph/badge.svg)](https://codecov.io/gh/adrianbrad/queue)

[![lint-test](https://github.com/adrianbrad/queue/actions/workflows/lint-test.yaml/badge.svg)](https://github.com/adrianbrad/queue/actions?query=workflow%3Alint-test)
[![grype](https://github.com/adrianbrad/queue/actions/workflows/grype.yaml/badge.svg)](https://github.com/adrianbrad/queue/actions?query=workflow%3Agrype)
[![codeql](https://github.com/adrianbrad/queue/actions/workflows/codeql.yaml/badge.svg)](https://github.com/adrianbrad/queue/actions?query=workflow%3ACodeQL)

---

Queue is a Go package providing different thread-safe generic queue implementations.

Available queues:
#### Blocking Queue
  - Waits for the queue have elements available before retrieving from it.
  - #### TODO:
    - [ ]  `Put()` - append an element to the back of the queue
    - [ ]  `Peek()` - retrieve but do not remove the head of the queue