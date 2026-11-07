// Package queue provides thread-safe, generic blocking, priority, circular
// (ring buffer), linked FIFO and delay queues for Go.
//
// Five implementations share one Queue[T] interface:
//
// A blocking queue, which provides methods that wait for the
// queue to have available elements when attempting to retrieve an element, and
// waits for a free slot when attempting to insert an element.
//
// A priority queue based on container/heap. Order is defined by a less
// function supplied at construction; the head of the queue is always the
// highest priority element.
//
// A circular queue (ring buffer), which uses a fixed-size slice as if it
// were connected end-to-end. When the queue is full, adding a new element
// overwrites the oldest element.
//
// A linked queue, an unbounded FIFO implemented as a singly linked list,
// offering O(1) time complexity for enqueue and dequeue operations. The queue
// maintains pointers to both the head (front) and tail (end) of the list for
// efficient operations without the need for traversal.
//
// A delay queue, in which each element becomes available at a deadline
// computed by a caller-supplied function at insertion time. Get errors until
// the head is due, and GetWait sleeps until it is.
package queue
