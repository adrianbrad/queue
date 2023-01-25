package queue

// Queue is a collection that orders elements in a FIFO order.
// This interface provides basic methods for adding and extracting elements
// from the queue.
// Items are extracted from the head of the queue and added to the tail
// of the queue.
type Queue[T any] interface {
	// Peek retrieves but does not remove the head of the queue.
	Peek() (T, error)

	// Size returns the number of elements in the queue.
	Size() int

	// Get retrieves and removes the head of the queue.
	Get() (T, error)

	// Offer inserts the element to the tail of the queue.
	Offer(T) error

	// Reset sets the queue to its initial state.
	Reset()
}
