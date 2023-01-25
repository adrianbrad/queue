package queue

import (
	"container/heap"
	"sort"
	"sync"
)

// Lesser is the interface that wraps basic Less method.
type Lesser interface {
	// Less compares the caller to the other
	Less(other any) bool
}

// Ensure Priority implements the heap.Interface.
var _ heap.Interface = (*priorityHeap[noopLesser])(nil)

// priorityHeap implements the heap.Interface, thus enabling this struct
// to be accepted as a parameter for the methods available in the heap package.
type priorityHeap[T Lesser] []T

// Len is the number of elements in the collection.
func (h *priorityHeap[T]) Len() int {
	return len(*h)
}

// Less reports whether the element with index i
// must sort before the element with index j.
func (h *priorityHeap[T]) Less(i, j int) bool {
	return (*h)[i].Less((*h)[j])
}

// Swap swaps the elements with indexes i and j.
func (h *priorityHeap[T]) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

// Push inserts elem into the heap.
func (h *priorityHeap[T]) Push(elem any) {
	//nolint: forcetypeassert // since priorityHeap is unexported, this
	// method cannot be directly called by a library client, it is only called
	// by the heap package functions. Thus, it is safe to expect that the
	// input parameter `elem` type is always T.
	*h = append(*h, elem.(T))
}

// Pop removes and returns the highest priority element.
func (h *priorityHeap[T]) Pop() any {
	n := len(*h)

	elem := (*h)[n-1]

	*h = (*h)[0 : n-1]

	return elem
}

// Ensure Priority implements the Queue interface.
var _ Queue[noopLesser] = (*Priority[noopLesser])(nil)

// Priority is a Queue implementation.
// ! The elements must implement the Lesser interface.
//
// The ordering is given by the Lesser.Less method implementation.
// The head of the queue is always the highest priority element.
//
// ! If capacity is provided and is less than the number of elements provided,
// the elements slice is sorted and trimmed to fit the capacity.
//
// For ordered types (types that support the operators < <= >= >), the order
// can be defined by using the following operators:
// > - for ascending order
// < - for descending order.
type Priority[T Lesser] struct {
	initialElements priorityHeap[T]
	elements        priorityHeap[T]

	capacity *int

	// synchronization
	lock sync.RWMutex
}

// NewPriority returns a new Priority Queue containing the given elements.
func NewPriority[T Lesser](
	elems []T,
	opts ...Option,
) *Priority[T] {
	// default options
	options := options{
		capacity: nil,
	}

	for _, o := range opts {
		o.apply(&options)
	}

	elementsHeap := priorityHeap[T](elems)

	// if capacity is provided and is less than the number of elements
	// provided, the elements are sorted and trimmed to fit the capacity.
	if options.capacity != nil && *options.capacity < len(elementsHeap) {
		sort.Slice(elementsHeap, func(i, j int) bool {
			return elementsHeap[i].Less(elementsHeap[j])
		})

		elementsHeap = elementsHeap[:*options.capacity]
	}

	heap.Init(&elementsHeap)

	initialElems := make(priorityHeap[T], len(elementsHeap))

	copy(initialElems, elementsHeap)

	return &Priority[T]{
		initialElements: initialElems,
		elements:        elementsHeap,
		capacity:        options.capacity,
	}
}

// ==================================Insertion=================================

// Offer inserts the element into the queue.
// If the queue is full it returns the ErrQueueIsFull error.
func (pq *Priority[T]) Offer(elem T) error {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	if pq.capacity != nil && len(pq.elements) >= *pq.capacity {
		return ErrQueueIsFull
	}

	heap.Push(&pq.elements, elem)

	return nil
}

// Reset sets the queue to its initial stat, by replacing the current
// elements with the elements provided at creation.
func (pq *Priority[T]) Reset() {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	if len(pq.elements) > len(pq.initialElements) {
		pq.elements = pq.elements[:len(pq.initialElements)]
	}

	if len(pq.elements) < len(pq.initialElements) {
		pq.elements = make([]T, len(pq.initialElements))
	}

	copy(pq.elements, pq.initialElements)
}

// ===================================Removal==================================

// Get removes and returns the head of the queue.
// If no element is available it returns an ErrNoElementsAvailable error.
func (pq *Priority[T]) Get() (elem T, _ error) {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	if len(pq.elements) == 0 {
		return elem, ErrNoElementsAvailable
	}

	//nolint: forcetypeassert // since the heap package does not yet support
	// generic types it has to use the `any` type. In this case, by design,
	// type of the items available in the pq.elements collection is always T.
	return heap.Pop(&pq.elements).(T), nil
}

// =================================Examination================================

// Peek retrieves but does not return the head of the queue.
func (pq *Priority[T]) Peek() (elem T, _ error) {
	pq.lock.RLock()
	defer pq.lock.RUnlock()

	if len(pq.elements) == 0 {
		return elem, ErrNoElementsAvailable
	}

	return pq.elements[0], nil
}

// Size returns the number of elements in the queue.
func (pq *Priority[T]) Size() int {
	pq.lock.RLock()
	defer pq.lock.RUnlock()

	return len(pq.elements)
}
