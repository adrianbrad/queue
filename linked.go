package queue

import (
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
}

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

	value := lq.head.value
	lq.head = lq.head.next
	lq.size--

	if lq.isEmpty() {
		lq.tail = nil
	}

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
	newNode := &node[T]{value: value}

	if lq.isEmpty() {
		lq.head = newNode
	} else {
		lq.tail.next = newNode
	}

	lq.tail = newNode
	lq.size++

	return nil
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

// IsEmpty returns true if the queue is empty, false otherwise.
func (lq *Linked[T]) isEmpty() bool {
	return lq.size == 0
}

// Iterator returns a channel that will be filled with the elements.
// It removes the elements from the queue.
func (lq *Linked[T]) Iterator() <-chan T {
	ch := make(chan T)

	elems := lq.Clear()

	go func() {
		for _, e := range elems {
			ch <- e
		}

		close(ch)
	}()

	return ch
}

// Clear removes and returns all elements from the queue.
func (lq *Linked[T]) Clear() []T {
	lq.lock.Lock()
	defer lq.lock.Unlock()

	elements := make([]T, 0, lq.size)

	current := lq.head
	for current != nil {
		elements = append(elements, current.value)
		next := current.next
		current = next
	}

	// Clear the queue
	lq.head = nil
	lq.tail = nil
	lq.size = 0

	return elements
}
