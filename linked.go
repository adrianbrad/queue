package queue

import (
	"encoding/json"
	"sync"
)

var _ Queue[any] = (*Linked[any])(nil)

// node is an individual element of the linked list.
type node[T any] struct {
	value T
	next  *node[T]
}

// Linked represents a data structure representing a queue that uses a
// linked list for its internal storage.
type Linked[T comparable] struct {
	head *node[T] // first node of the queue.
	tail *node[T] // last node of the queue.
	size int      // number of elements in the queue.
	// nolint: revive
	initialElements []T // initial elements with which the queue was created, allowing for a reset to its original state if needed.
	// synchronization
	lock sync.RWMutex
	// free is a stack of recycled nodes. Offer pulls from here before
	// allocating; Get pushes onto it. Bounded to freeCap so Clear/Reset
	// can't cause unbounded retention.
	free    *node[T]
	freeLen int
}

// freeCap is the maximum number of nodes cached for reuse.
const freeCap = 64

// NewLinked creates a new Linked containing the given elements.
func NewLinked[T comparable](elements []T) *Linked[T] {
	queue := &Linked[T]{
		head:            nil,
		tail:            nil,
		size:            0,
		initialElements: make([]T, len(elements)),
	}

	copy(queue.initialElements, elements)

	for _, element := range elements {
		_ = queue.offer(element)
	}

	return queue
}

// Get retrieves and removes the head of the queue.
func (lq *Linked[T]) Get() (elem T, _ error) {
	lq.lock.Lock()
	defer lq.lock.Unlock()

	if lq.isEmpty() {
		return elem, ErrNoElementsAvailable
	}

	popped := lq.head
	value := popped.value

	lq.head = popped.next
	lq.size--

	if lq.isEmpty() {
		lq.tail = nil
	}

	lq.recycle(popped)

	return value, nil
}

// Offer inserts the element into the queue.
func (lq *Linked[T]) Offer(value T) error {
	lq.lock.Lock()
	defer lq.lock.Unlock()

	return lq.offer(value)
}

// offer inserts the element into the queue.
func (lq *Linked[T]) offer(value T) error {
	newNode := lq.acquireNode()
	newNode.value = value

	if lq.isEmpty() {
		lq.head = newNode
	} else {
		lq.tail.next = newNode
	}

	lq.tail = newNode
	lq.size++

	return nil
}

// acquireNode pulls a node off the free list or allocates a fresh one.
func (lq *Linked[T]) acquireNode() *node[T] {
	if lq.free == nil {
		return &node[T]{}
	}

	n := lq.free
	lq.free = n.next
	lq.freeLen--
	n.next = nil

	return n
}

// recycle zeroes a popped node and returns it to the free list, capped
// at freeCap to avoid pinning memory after a large drain.
func (lq *Linked[T]) recycle(n *node[T]) {
	if lq.freeLen >= freeCap {
		return
	}

	var zero T

	n.value = zero
	n.next = lq.free
	lq.free = n
	lq.freeLen++
}

// Reset sets the queue to its initial state.
func (lq *Linked[T]) Reset() {
	lq.lock.Lock()
	defer lq.lock.Unlock()

	lq.head = nil
	lq.tail = nil
	lq.size = 0

	for _, element := range lq.initialElements {
		_ = lq.offer(element)
	}
}

// Contains returns true if the queue contains the element.
func (lq *Linked[T]) Contains(value T) bool {
	lq.lock.RLock()
	defer lq.lock.RUnlock()

	current := lq.head
	for current != nil {
		if current.value == value {
			return true
		}

		current = current.next
	}

	return false
}

// Peek retrieves but does not remove the head of the queue.
func (lq *Linked[T]) Peek() (elem T, _ error) {
	lq.lock.RLock()
	defer lq.lock.RUnlock()

	if lq.isEmpty() {
		return elem, ErrNoElementsAvailable
	}

	return lq.head.value, nil
}

// Size returns the number of elements in the queue.
func (lq *Linked[T]) Size() int {
	lq.lock.RLock()
	defer lq.lock.RUnlock()

	return lq.size
}

// IsEmpty returns true if the queue is empty, false otherwise.
func (lq *Linked[T]) IsEmpty() bool {
	lq.lock.RLock()
	defer lq.lock.RUnlock()

	return lq.isEmpty()
}

// isEmpty returns true if the queue is empty, false otherwise.
func (lq *Linked[T]) isEmpty() bool {
	return lq.size == 0
}

// Iterator returns a channel that will be filled with the elements.
// It removes the elements from the queue.
func (lq *Linked[T]) Iterator() <-chan T {
	lq.lock.Lock()
	defer lq.lock.Unlock()

	elems := lq.drainLocked()

	ch := make(chan T, len(elems))

	for i := range elems {
		ch <- elems[i]
	}

	close(ch)

	return ch
}

// Clear removes and returns all elements from the queue.
func (lq *Linked[T]) Clear() []T {
	lq.lock.Lock()
	defer lq.lock.Unlock()

	return lq.drainLocked()
}

// drainLocked collects all elements in order and resets the queue.
// Caller must hold the write lock.
func (lq *Linked[T]) drainLocked() []T {
	elements := make([]T, lq.size)

	current := lq.head
	for i := 0; current != nil; i++ {
		elements[i] = current.value
		current = current.next
	}

	lq.head = nil
	lq.tail = nil
	lq.size = 0

	return elements
}

// MarshalJSON serializes the Linked queue to JSON.
func (lq *Linked[T]) MarshalJSON() ([]byte, error) {
	lq.lock.RLock()
	defer lq.lock.RUnlock()

	elements := make([]T, lq.size)

	current := lq.head
	for i := 0; current != nil; i++ {
		elements[i] = current.value
		current = current.next
	}

	return json.Marshal(elements)
}
