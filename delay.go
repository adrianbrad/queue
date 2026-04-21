package queue

import (
	"encoding/json"
	"sort"
	"sync"
	"time"
)

// delayed pairs an element with its cached deadline.
type delayed[T any] struct {
	elem     T
	deadline time.Time
}

// delayHeap is a min-heap over delayed[T] keyed by deadline.
//
// The heap algorithm is implemented directly on the typed slice instead
// of via container/heap to avoid boxing delayed[T] into `any` on every
// push and pop — the dominant allocation source in the benchmark profile.
type delayHeap[T any] struct {
	items []delayed[T]
}

func (h *delayHeap[T]) len() int { return len(h.items) }

func (h *delayHeap[T]) less(i, j int) bool {
	return h.items[i].deadline.Before(h.items[j].deadline)
}

func (h *delayHeap[T]) swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
}

// push appends x and restores the heap invariant.
func (h *delayHeap[T]) push(x delayed[T]) {
	h.items = append(h.items, x)
	h.up(len(h.items) - 1)
}

// pop removes and returns the root (earliest-deadline) element.
// The queue must be non-empty.
func (h *delayHeap[T]) pop() delayed[T] {
	n := len(h.items) - 1

	h.swap(0, n)
	h.down(0, n)

	x := h.items[n]

	var zero delayed[T]

	h.items[n] = zero
	h.items = h.items[:n]

	return x
}

// up sifts the item at index i toward the root.
func (h *delayHeap[T]) up(i int) {
	for {
		parent := (i - 1) / 2 //nolint:mnd // standard binary-heap parent index.
		if parent == i || !h.less(i, parent) {
			return
		}

		h.swap(parent, i)
		i = parent
	}
}

// down sifts the item at index i toward the leaves; n bounds the
// active heap region.
func (h *delayHeap[T]) down(i, n int) {
	for {
		left := 2*i + 1 //nolint:mnd // standard binary-heap left-child index.
		if left >= n || left < 0 {
			return
		}

		j := left
		if right := left + 1; right < n && h.less(right, left) {
			j = right
		}

		if !h.less(j, i) {
			return
		}

		h.swap(i, j)
		i = j
	}
}

// Ensure Delay implements the Queue interface.
var _ Queue[any] = (*Delay[any])(nil)

// Delay is a Queue implementation where each element becomes dequeuable
// at a deadline computed by a caller-supplied function at Offer time.
//
// Get returns ErrNoElementsAvailable if the queue is empty or the head's
// deadline has not yet passed; GetWait sleeps until the head becomes due.
type Delay[T comparable] struct {
	deadlineFunc func(T) time.Time
	items        *delayHeap[T]
	initial      []T
	capacity     *int

	lock     sync.Mutex
	notEmpty *sync.Cond
}

// NewDelay creates a Delay queue. deadlineFunc is called at Offer and
// Reset time to compute each element's deadline; the deadline is cached
// per element (not re-evaluated on every Get).
// Panics if deadlineFunc is nil or WithCapacity is negative.
func NewDelay[T comparable](
	elems []T,
	deadlineFunc func(T) time.Time,
	opts ...Option,
) *Delay[T] {
	if deadlineFunc == nil {
		panic("nil deadline func")
	}

	options := options{capacity: nil}

	for _, o := range opts {
		o.apply(&options)
	}

	if options.capacity != nil && *options.capacity < 0 {
		panic("negative capacity")
	}

	effective := elems
	if options.capacity != nil && *options.capacity < len(effective) {
		effective = effective[:*options.capacity]
	}

	initial := make([]T, len(effective))
	copy(initial, effective)

	dq := &Delay[T]{
		deadlineFunc: deadlineFunc,
		items:        &delayHeap[T]{items: make([]delayed[T], 0, len(effective))},
		initial:      initial,
		capacity:     options.capacity,
	}

	dq.notEmpty = sync.NewCond(&dq.lock)

	for _, e := range effective {
		dq.items.push(delayed[T]{elem: e, deadline: deadlineFunc(e)})
	}

	return dq
}

// ==================================Insertion=================================

// Offer inserts elem with deadline = deadlineFunc(elem).
// Returns ErrQueueIsFull when constructed WithCapacity and already at limit.
func (dq *Delay[T]) Offer(elem T) error {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	if dq.capacity != nil && dq.items.len() >= *dq.capacity {
		return ErrQueueIsFull
	}

	dq.items.push(delayed[T]{
		elem:     elem,
		deadline: dq.deadlineFunc(elem),
	})

	dq.notEmpty.Broadcast()

	return nil
}

// Reset restores the queue to the elements provided at construction,
// recomputing their deadlines with the original deadlineFunc.
func (dq *Delay[T]) Reset() {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	dq.items.items = dq.items.items[:0]

	for _, e := range dq.initial {
		dq.items.push(delayed[T]{
			elem:     e,
			deadline: dq.deadlineFunc(e),
		})
	}

	dq.notEmpty.Broadcast()
}

// ===================================Removal==================================

// Get returns the head if its deadline has passed, otherwise
// ErrNoElementsAvailable. Never blocks.
func (dq *Delay[T]) Get() (v T, _ error) {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	if dq.items.len() == 0 {
		return v, ErrNoElementsAvailable
	}

	if time.Now().Before(dq.items.items[0].deadline) {
		return v, ErrNoElementsAvailable
	}

	elem := dq.items.pop().elem

	dq.notEmpty.Broadcast()

	return elem, nil
}

// GetWait blocks until the head's deadline passes and returns that
// element. If the queue is empty, waits for an Offer.
func (dq *Delay[T]) GetWait() T {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	for {
		if dq.items.len() > 0 {
			now := time.Now()
			if !now.Before(dq.items.items[0].deadline) {
				return dq.items.pop().elem
			}

			// Head is not yet due: schedule a timer that Broadcasts when
			// the deadline passes, then Wait. Any earlier Offer / Reset /
			// Clear also Broadcasts, so we re-check on state changes too.
			remaining := dq.items.items[0].deadline.Sub(now)
			timer := time.AfterFunc(remaining, func() {
				dq.lock.Lock()
				dq.notEmpty.Broadcast()
				dq.lock.Unlock()
			})

			dq.notEmpty.Wait()
			timer.Stop()

			continue
		}

		dq.notEmpty.Wait()
	}
}

// Clear removes and returns all elements in deadline order.
func (dq *Delay[T]) Clear() []T {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	n := dq.items.len()
	out := make([]T, n)

	for i := 0; i < n; i++ {
		out[i] = dq.items.pop().elem
	}

	dq.notEmpty.Broadcast()

	return out
}

// Iterator returns a channel that receives all elements in deadline
// order. Elements are removed from the queue.
func (dq *Delay[T]) Iterator() <-chan T {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	ch := make(chan T, dq.items.len())

	for dq.items.len() > 0 {
		ch <- dq.items.pop().elem
	}

	close(ch)

	dq.notEmpty.Broadcast()

	return ch
}

// =================================Examination================================

// Peek returns the head regardless of whether its deadline has passed.
// Returns ErrNoElementsAvailable if the queue is empty.
func (dq *Delay[T]) Peek() (v T, _ error) {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	if dq.items.len() == 0 {
		return v, ErrNoElementsAvailable
	}

	return dq.items.items[0].elem, nil
}

// Size returns the number of elements in the queue, due or not.
func (dq *Delay[T]) Size() int {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	return dq.items.len()
}

// IsEmpty returns true if the queue contains no elements.
func (dq *Delay[T]) IsEmpty() bool {
	return dq.Size() == 0
}

// Contains reports whether the given element is in the queue.
func (dq *Delay[T]) Contains(elem T) bool {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	for i := range dq.items.items {
		if dq.items.items[i].elem == elem {
			return true
		}
	}

	return false
}

// MarshalJSON serializes the Delay queue to JSON in deadline order.
func (dq *Delay[T]) MarshalJSON() ([]byte, error) {
	dq.lock.Lock()

	snapshot := make([]delayed[T], len(dq.items.items))
	copy(snapshot, dq.items.items)

	dq.lock.Unlock()

	sort.Slice(snapshot, func(i, j int) bool {
		return snapshot[i].deadline.Before(snapshot[j].deadline)
	})

	out := make([]T, len(snapshot))
	for i := range snapshot {
		out[i] = snapshot[i].elem
	}

	return json.Marshal(out)
}
