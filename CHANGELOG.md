## v1.5.0 (2026-04-20)

### Fix

- **priority,blocking,circular**: `Iterator()` now takes a write lock — it mutates queue state (was: read lock, race-prone)
- **circular**: `Contains` loop bounds with non-zero head no longer skips elements or reads out of range
- **circular**: `drainQueue` drains all elements (loop size captured before `pop` decrements it)

### Perf

- **circular**: preallocate slice for `Clear` and `MarshalJSON`; zero popped slot for GC
- **linked**: `Iterator` fills a buffered channel synchronously instead of spawning a goroutine
- **linked**: preallocate `Clear` and `MarshalJSON` results
- **priority**: `Reset` simplified to single reslice + copy

### Docs / security

- Add `SECURITY.md` with private-disclosure process
- Add `CONTRIBUTING.md`
- Add `.github/CODEOWNERS`
- Harden all GitHub Actions workflows: pinned by commit SHA, minimum `permissions:` scope
- Add OpenSSF Scorecard workflow + README badges
- Enforce 100% test coverage in CI
- Add signed releases via goreleaser + cosign keyless (Sigstore)

## v1.4.0 (2025-05-13)

### Feat

- **blocking**: fix `Reset()` unbound memory growth potential
- implement `MarshalJSON` for all queues

## v1.3.0 (2023-10-24)

### Feat

- **linked**: add linked queue implementation with tests

### Fix

- **linked**: make linked queue implementation thread-safe

## v1.2.1 (2023-05-08)

### Fix

- fix a bug caused by not calling the (*sync.Cond).Wait() method in a loop

## v1.2.0 (2023-04-14)

### Feat

- implement `Circular Queue` and add tests

## v1.1.0 (2023-03-23)

### Feat

- change the queue type parameter from `any` to `comparable`
- add `Contains`, `Peek`, `Size`, `IsEmpty`, `Iterator` and `Clear` methods to the Queue interface and implementations. In order to implement the `Contains` method the type parameter used by the queues was changed to `comparable` from `any`

## v0.9.0 (2023-02-03)

### Feat

- **priority**: removed lesser interface. added fuzz and benchmarks
- **priority**: remove lesser interface and use a less func

## v0.8.0 (2023-01-25)

### Feat

- Implement reset for all queue. Add missing source code from previous commit. Update readme
- add missing source code. update readme
- improve the `Queue` interface, implement `Priority` queue, fix `Blocking` tests, provide example tests

## v0.7.0 (2023-01-13)

### Feat

- **blocking**: implement the `Offer` method which returns an error in case the queue is full

## v0.6.0 (2023-01-12)

### Feat

- **options**: added `capacity` option
- **blocking**: implement `Get` method, it returns an error if there are no available elements in the queue

## v0.5.0 (2023-01-08)

### Feat

- **blocking**: remove internal channel and implement peek

## v0.4.1 (2022-09-12)

### Refactor

- **blocking**: remove unnecessary ctx.Done case from Blocking.Refill()

## v0.4.0 (2022-09-11)

### Feat

- **blocking**: use buffered channel as the queue storage, as u/Cidan suggested

## v0.3.1 (2022-09-03)

### Refactor

- **blocking**: change the index type from `atomic.Uint32` to `atomic.Uintptr`

## v0.3.0 (2022-09-02)

### Feat

- **blocking**: add `Peek` method, which returns but does not remove an element

### Refactor

- **blocking**: store index and sync as values instead of pointers

## v0.2.0 (2022-09-01)

### Feat

- **blocking**: as per u/skeeto comment, remove the useless error returns

### Fix

- **deadlock**: fix the deadlock caused by unsynchronized index and broadcast channel

## v0.1.1 (2022-09-01)

### Refactor

- **blocking_queue**: rename `Get` into `Take`

## v0.1.0 (2022-09-01)

### Feat

- **blocking_queue**: add first `blocking queue` implementation
