package queue

var _ Queue[any] = (*Linked[any])(nil)

// node is an individual element of the linked list.
type node[T any] struct {
	value T
	next  *node[T]
}

// Linked represents a data structure representing a queue that uses a
// linked list for its internal storage.
// ! The Linked Queue is not thread safe.
type Linked[T comparable] struct {
	head *node[T] // first node of the queue.
	tail *node[T] // last node of the queue.
	size int      // number of elements in the queue.
	// nolint: revive
	initialElements []T // initial elements with which the queue was created, allowing for a reset to its original state if needed.
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
		_ = queue.Offer(element)
	}

	return queue
}

// Get retrieves and removes the head of the queue.
func (q *Linked[T]) Get() (elem T, _ error) {
	if q.IsEmpty() {
		return elem, ErrNoElementsAvailable
	}

	value := q.head.value
	q.head = q.head.next
	q.size--

	if q.IsEmpty() {
		q.tail = nil
	}

	return value, nil
}

// Offer inserts the element into the queue.
func (q *Linked[T]) Offer(value T) error {
	newNode := &node[T]{value: value}

	if q.IsEmpty() {
		q.head = newNode
	} else {
		q.tail.next = newNode
	}

	q.tail = newNode
	q.size++

	return nil
}

// Reset sets the queue to its initial state.
func (q *Linked[T]) Reset() {
	q.head = nil
	q.tail = nil
	q.size = 0

	for _, element := range q.initialElements {
		_ = q.Offer(element)
	}
}

// Contains returns true if the queue contains the element.
func (q *Linked[T]) Contains(value T) bool {
	current := q.head
	for current != nil {
		if current.value == value {
			return true
		}

		current = current.next
	}

	return false
}

// Peek retrieves but does not remove the head of the queue.
func (q *Linked[T]) Peek() (elem T, _ error) {
	if q.IsEmpty() {
		return elem, ErrNoElementsAvailable
	}

	return q.head.value, nil
}

// Size returns the number of elements in the queue.
func (q *Linked[T]) Size() int {
	return q.size
}

// IsEmpty returns true if the queue is empty, false otherwise.
func (q *Linked[T]) IsEmpty() bool {
	return q.size == 0
}

// Iterator returns a channel that will be filled with the elements.
// It removes the elements from the queue.
func (q *Linked[T]) Iterator() <-chan T {
	ch := make(chan T)

	elems := q.Clear()

	go func() {
		for _, e := range elems {
			ch <- e
		}

		close(ch)
	}()

	return ch
}

// Clear removes and returns all elements from the queue.
func (q *Linked[T]) Clear() []T {
	elements := make([]T, 0, q.size)

	current := q.head
	for current != nil {
		elements = append(elements, current.value)
		next := current.next
		current = next
	}

	// Clear the queue
	q.head = nil
	q.tail = nil
	q.size = 0

	return elements
}
