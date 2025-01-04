package queue

import (
	"container/heap"
	"encoding/json"
	"sort"
	"sync"
)

// Ensure Priority implements the heap.Interface.
var _ heap.Interface = (*priorityHeap[any])(nil)

// priorityHeap implements the heap.Interface, thus enabling this struct
// to be accepted as a parameter for the methods available in the heap package.
type priorityHeap[T comparable] struct {
	elems    []T
	lessFunc func(elem, otherElem T) bool
}

// Len is the number of elements in the collection.
func (h *priorityHeap[T]) Len() int {
	return len(h.elems)
}

// Less reports whether the element with index i
// must sort before the element with index j.
func (h *priorityHeap[T]) Less(i, j int) bool {
	return h.lessFunc(h.elems[i], h.elems[j])
}

// Swap swaps the elements with indexes i and j.
func (h *priorityHeap[T]) Swap(i, j int) {
	h.elems[i], h.elems[j] = h.elems[j], h.elems[i]
}

// Push inserts elem into the heap.
func (h *priorityHeap[T]) Push(elem any) {
	// nolint: forcetypeassert // since priorityHeap is unexported, this
	// method cannot be directly called by a library client, it is only called
	// by the heap package functions. Thus, it is safe to expect that the
	// input parameter `elem` type is always T.
	h.elems = append(h.elems, elem.(T))
}

// Pop removes and returns the highest priority element.
func (h *priorityHeap[T]) Pop() any {
	n := len(h.elems)

	elem := (h.elems)[n-1]

	h.elems = (h.elems)[0 : n-1]

	return elem
}

// Ensure Priority implements the Queue interface.
var _ Queue[any] = (*Priority[any])(nil)

// Priority is a Queue implementation.
//
// The ordering is given by the lessFunc.
// The head of the queue is always the highest priority element.
//
// ! If capacity is provided and is less than the number of elements provided,
// the elements slice is sorted and trimmed to fit the capacity.
//
// For ordered types (types that support the operators < <= >= >), the order
// can be defined by using the following operators:
// > - for ascending order
// < - for descending order.
type Priority[T comparable] struct {
	initialElements []T
	elements        *priorityHeap[T]

	capacity *int

	// synchronization
	lock sync.RWMutex
}

// NewPriority creates a new Priority Queue containing the given elements.
// It panics if lessFunc is nil.
func NewPriority[T comparable](
	elems []T,
	lessFunc func(elem, otherElem T) bool,
	opts ...Option,
) *Priority[T] {
	if lessFunc == nil {
		panic("nil less func")
	}

	// default options
	options := options{
		capacity: nil,
	}

	for _, o := range opts {
		o.apply(&options)
	}

	heapElems := make([]T, len(elems))

	copy(heapElems, elems)

	elementsHeap := &priorityHeap[T]{
		elems:    heapElems,
		lessFunc: lessFunc,
	}

	// if capacity is provided and is less than the number of elements
	// provided, the elements are sorted and trimmed to fit the capacity.
	if options.capacity != nil && *options.capacity < elementsHeap.Len() {
		sort.Slice(elementsHeap.elems, func(i, j int) bool {
			return lessFunc((elementsHeap.elems)[i], (elementsHeap.elems)[j])
		})

		elementsHeap.elems = (elementsHeap.elems)[:*options.capacity]
	}

	heap.Init(elementsHeap)

	initialElems := make([]T, elementsHeap.Len())

	copy(initialElems, elementsHeap.elems)

	pq := &Priority[T]{
		initialElements: initialElems,
		elements:        elementsHeap,
		capacity:        options.capacity,
	}

	return pq
}

// ==================================Insertion=================================

// Offer inserts the element into the queue.
// If the queue is full it returns the ErrQueueIsFull error.
func (pq *Priority[T]) Offer(elem T) error {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	if pq.capacity != nil && pq.elements.Len() >= *pq.capacity {
		return ErrQueueIsFull
	}

	heap.Push(pq.elements, elem)

	return nil
}

// Reset sets the queue to its initial stat, by replacing the current
// elements with the elements provided at creation.
func (pq *Priority[T]) Reset() {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	if pq.elements.Len() > len(pq.initialElements) {
		pq.elements.elems = (pq.elements.elems)[:len(pq.initialElements)]
	}

	if pq.elements.Len() < len(pq.initialElements) {
		pq.elements.elems = make([]T, len(pq.initialElements))
	}

	copy(pq.elements.elems, pq.initialElements)
}

// ===================================Removal==================================

// Get removes and returns the head of the queue.
// If no element is available it returns an ErrNoElementsAvailable error.
func (pq *Priority[T]) Get() (elem T, _ error) {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	if pq.elements.Len() == 0 {
		return elem, ErrNoElementsAvailable
	}

	// nolint: forcetypeassert, revive // since the heap package does not yet support
	// generic types it has to use the `any` type. In this case, by design,
	// type of the items available in the pq.elements collection is always T.
	return heap.Pop(pq.elements).(T), nil
}

// Clear removes all elements from the queue.
func (pq *Priority[T]) Clear() []T {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	elemsLen := pq.elements.Len()

	elems := make([]T, elemsLen)

	for i := 0; i < elemsLen; i++ {
		// nolint: forcetypeassert, revive // since priorityHeap is unexported, this
		// method cannot be directly called by a library client, it is only called
		// by the heap package functions. Thus, it is safe to expect that the
		// input parameter `elem` type is always T.
		elems[i] = heap.Pop(pq.elements).(T)
	}

	return elems
}

// Iterator returns an iterator over the elements in the queue.
// It removes the elements from the queue.
func (pq *Priority[T]) Iterator() <-chan T {
	pq.lock.RLock()
	defer pq.lock.RUnlock()

	// use a buffered channel to avoid blocking the iterator.
	iteratorCh := make(chan T, pq.elements.Len())

	// iterate over the elements and send them to the channel.
	for pq.elements.Len() > 0 {
		// nolint: forcetypeassert, revive // since priorityHeap is unexported, this
		// method cannot be directly called by a library client, it is only called
		// by the heap package functions. Thus, it is safe to expect that the
		// input parameter `elem` type is always T.
		iteratorCh <- heap.Pop(pq.elements).(T)
	}

	close(iteratorCh)

	return iteratorCh
}

// =================================Examination================================

// IsEmpty returns true if the queue is empty, false otherwise.
func (pq *Priority[T]) IsEmpty() bool {
	pq.lock.RLock()
	defer pq.lock.RUnlock()

	return pq.elements.Len() == 0
}

// Contains returns true if the queue contains the element, false otherwise.
func (pq *Priority[T]) Contains(a T) bool {
	pq.lock.RLock()
	defer pq.lock.RUnlock()

	for i := range pq.elements.elems {
		if pq.elements.elems[i] == a {
			return true
		}
	}

	return false
}

// Peek retrieves but does not return the head of the queue.
func (pq *Priority[T]) Peek() (elem T, _ error) {
	pq.lock.RLock()
	defer pq.lock.RUnlock()

	if pq.elements.Len() == 0 {
		return elem, ErrNoElementsAvailable
	}

	return pq.elements.elems[0], nil
}

// Size returns the number of elements in the queue.
func (pq *Priority[T]) Size() int {
	pq.lock.RLock()
	defer pq.lock.RUnlock()

	return pq.elements.Len()
}

// MarshalJSON serializes the Priority queue to JSON.
func (pq *Priority[T]) MarshalJSON() ([]byte, error) {
	pq.lock.RLock()

	// Create a temporary copy of the heap to extract elements in order.
	tempHeap := &priorityHeap[T]{
		elems:    make([]T, len(pq.elements.elems)),
		lessFunc: pq.elements.lessFunc,
	}

	copy(tempHeap.elems, pq.elements.elems)

	pq.lock.RUnlock()

	heap.Init(tempHeap)

	output := make([]T, len(tempHeap.elems))

	i := 0

	for tempHeap.Len() > 0 {
		// nolint: forcetypeassert, revive
		output[i] = tempHeap.Pop().(T)
		i++
	}

	return json.Marshal(output)
}
