// Package queue provides multiple thread-safe generic queue implementations.
// Currently, there are 2 available implementations:
//
// A blocking queue, which provides methods that wait for the
// queue to have available elements when attempting to retrieve an element, and
// waits for a free slot when attempting to insert an element.
//
// A priority queue based on a container.Heap. The elements in the queue
// must implement the Lesser interface, and are ordered based on the
// Less method. The head of the queue is always the highest priority element.
//
// A circular queue, which is a queue that uses a fixed-size slice as
// if it were connected end-to-end. When the queue is full, adding a new element to the queue
// overwrites the oldest element.
package queue
