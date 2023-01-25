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
