package queue_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/adrianbrad/queue"
)

func TestBlocking(t *testing.T) {
	t.Parallel()

	t.Run("Consistency", testBlockingConsistency)
	t.Run("Clear", testBlockingClear)
	t.Run("Contains", testBlockingContains)
	t.Run("Iterator", testBlockingIterator)
	t.Run("IsEmpty", testBlockingIsEmpty)
	t.Run("Reset", testBlockingReset)
	t.Run("OfferWait", testBlockingOfferWait)
	t.Run("Offer", testBlockingOffer)
	t.Run("Peek", testBlockingPeek)
	t.Run("PeekWait", testBlockingPeekWait)
	t.Run("Get", testBlockingGet)
	t.Run("WithCapacity", testBlockingWithCapacity)
	t.Run("CondWaitWithCapacity", testBlockingCondWaitWithCapacity)
	t.Run("MarshalJSON", testBlockingMarshalJSON)
}

func testBlockingConsistency(t *testing.T) {
	t.Parallel()

	t.Run("SequentialIteration", func(t *testing.T) {
		t.Parallel()

		elems := []int{1, 2, 3}

		blockingQueue := queue.NewBlocking(elems)

		for j := range elems {
			elem := blockingQueue.GetWait()

			if elems[j] != elem {
				t.Fatalf("expected elem to be %d, got %d", elems[j], elem)
			}
		}
	})

	t.Run("100ConcurrentGoroutinesReading", func(t *testing.T) {
		t.Parallel()

		const lenElements = 100

		ids := make([]int, lenElements)

		for i := 1; i <= lenElements; i++ {
			ids[i-1] = i
		}

		blockingQueue := queue.NewBlocking(ids)

		var (
			wg          sync.WaitGroup
			resultMutex sync.Mutex
		)

		wg.Add(lenElements)

		result := make([]int, 0, lenElements)

		for i := 0; i < lenElements; i++ {
			go func() {
				elem := blockingQueue.GetWait()

				resultMutex.Lock()

				result = append(result, elem)
				resultMutex.Unlock()

				defer wg.Done()
			}()
		}

		wg.Wait()

		sort.SliceStable(result, func(i, j int) bool {
			return result[i] < result[j]
		})

		if !reflect.DeepEqual(ids, result) {
			t.Fatalf("expected result to be %v, got %v", ids, result)
		}
	})

	t.Run("PeekWaitAndPushWaiting", func(t *testing.T) {
		t.Parallel()

		elems := []int{1}

		blockingQueue := queue.NewBlocking(elems)

		_ = blockingQueue.GetWait()

		var wg sync.WaitGroup

		wg.Add(2)

		peekDone := make(chan struct{})

		blockingQueue.Reset()

		go func() {
			defer wg.Done()
			defer close(peekDone)

			elem := blockingQueue.PeekWait()

			t.Log("peek done")

			if elems[0] != elem {
				t.Errorf("expected elem to be %d, got %d", elems[0], elem)
			}
		}()

		go func() {
			defer wg.Done()

			<-peekDone

			elem := blockingQueue.GetWait()
			if elems[0] != elem {
				t.Errorf("expected elem to be %d, got %d", elems[0], elem)
			}
		}()

		wg.Wait()
	})

	t.Run("ResetWhileMoreRoutinesThanElementsAreWaiting", func(t *testing.T) {
		t.Parallel()

		elems := []int{1, 2, 3}

		const noRoutines = 100

		for i := 1; i <= noRoutines; i++ {
			t.Run(
				fmt.Sprintf("%dRoutinesWaiting", i),
				func(t *testing.T) {
					testResetOnMultipleRoutinesFunc[int](elems, i)(t)
				},
			)
		}
	})
}

func testBlockingClear(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		elems := []int{1, 2, 3}

		blockingQueue := queue.NewBlocking(elems)

		queueElems := blockingQueue.Clear()

		if !reflect.DeepEqual(elems, queueElems) {
			t.Fatalf("expected elements to be %v, got %v", elems, queueElems)
		}
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()

		blockingQueue := queue.NewBlocking([]int{})

		queueElems := blockingQueue.Clear()

		if len(queueElems) != 0 {
			t.Fatalf("expected elements to be empty, got %v", queueElems)
		}
	})
}

func testBlockingContains(t *testing.T) {
	t.Parallel()

	t.Run("True", func(t *testing.T) {
		t.Parallel()

		elems := []int{1, 2, 3}

		blockingQueue := queue.NewBlocking(elems)

		if !blockingQueue.Contains(2) {
			t.Fatal("expected queue to contain 2")
		}
	})

	t.Run("False", func(t *testing.T) {
		t.Parallel()

		elems := []int{1, 2, 3}

		blockingQueue := queue.NewBlocking(elems)

		if blockingQueue.Contains(4) {
			t.Fatal("expected queue to not contain 4")
		}
	})
}

func testBlockingIterator(t *testing.T) {
	t.Parallel()

	elems := []int{1, 2, 3}

	blockingQueue := queue.NewBlocking(elems)

	iterCh := blockingQueue.Iterator()

	if !blockingQueue.IsEmpty() {
		t.Fatal("expected queue to be empty")
	}

	iterElems := make([]int, 0, len(elems))

	for e := range iterCh {
		iterElems = append(iterElems, e)
	}

	if !reflect.DeepEqual(elems, iterElems) {
		t.Fatalf("expected elements to be %v, got %v", elems, iterElems)
	}
}

func testBlockingIsEmpty(t *testing.T) {
	t.Parallel()

	t.Run("True", func(t *testing.T) {
		t.Parallel()

		blockingQueue := queue.NewBlocking([]int{})

		if !blockingQueue.IsEmpty() {
			t.Fatal("expected queue to be empty")
		}
	})

	t.Run("False", func(t *testing.T) {
		t.Parallel()

		blockingQueue := queue.NewBlocking([]int{1})

		if blockingQueue.IsEmpty() {
			t.Fatal("expected queue to not be empty")
		}
	})
}

func testBlockingReset(t *testing.T) {
	t.Parallel()

	t.Run("WithCapacity", func(t *testing.T) {
		t.Parallel()

		elems := []int{1, 2, 3}

		initialSize := len(elems)

		blockingQueue := queue.NewBlocking(
			elems,
			queue.WithCapacity(initialSize+1),
		)

		if blockingQueue.Size() != initialSize {
			t.Fatalf("expected size to be %d, got %d", initialSize, blockingQueue.Size())
		}

		blockingQueue.OfferWait(4)

		if blockingQueue.Size() != initialSize+1 {
			t.Fatalf("expected size to be %d, got %d", initialSize+1, blockingQueue.Size())
		}

		blockingQueue.Reset()

		if blockingQueue.Size() != initialSize {
			t.Fatalf("expected size to be %d, got %d", initialSize, blockingQueue.Size())
		}

		_ = blockingQueue.Clear()

		elem := make(chan int)

		go func() {
			elem <- blockingQueue.GetWait()
		}()

		blockingQueue.OfferWait(5)

		if e := <-elem; e != 5 {
			t.Fatalf("expected elem to be %d, got %d", 5, e)
		}
	})

	t.Run("UnblocksProducersWaitingOnCapacity", func(t *testing.T) {
		t.Parallel()

		// Start with fewer elements than capacity so Reset frees a slot.
		blockingQueue := queue.NewBlocking(
			[]int{1, 2},
			queue.WithCapacity(3),
		)

		if err := blockingQueue.Offer(3); err != nil {
			t.Fatalf("unexpected err filling queue: %v", err)
		}

		offerReturned := make(chan struct{})

		go func() {
			blockingQueue.OfferWait(4)
			close(offerReturned)
		}()

		// Give the producer a moment to enter notFullCond.Wait().
		time.Sleep(10 * time.Millisecond)

		select {
		case <-offerReturned:
			t.Fatal("OfferWait returned before Reset; queue was full")
		default:
		}

		// Reset shrinks back to the initial two elements, leaving a free slot.
		// Without broadcasting notFullCond, the producer never wakes.
		blockingQueue.Reset()

		select {
		case <-offerReturned:
		case <-time.After(time.Second):
			t.Fatal("OfferWait was not unblocked by Reset")
		}
	})

	t.Run("InitialElemsExceedCapacity", func(t *testing.T) {
		t.Parallel()

		blockingQueue := queue.NewBlocking(
			[]int{1, 2, 3, 4, 5},
			queue.WithCapacity(3),
		)

		if size := blockingQueue.Size(); size != 3 {
			t.Fatalf("expected size to be 3 after construction, got %d", size)
		}

		// Drain the queue then Reset. Reset must honour capacity rather
		// than restoring the original slice verbatim.
		_ = blockingQueue.Clear()

		blockingQueue.Reset()

		if size := blockingQueue.Size(); size != 3 {
			t.Fatalf("expected size to be 3 after Reset, got %d", size)
		}

		if err := blockingQueue.Offer(99); !errors.Is(err, queue.ErrQueueIsFull) {
			t.Fatalf("expected ErrQueueIsFull after Reset fills capacity, got %v", err)
		}
	})
}

func testBlockingOfferWait(t *testing.T) {
	t.Parallel()

	t.Run("NoCapacity", func(t *testing.T) {
		t.Parallel()

		elems := []int{1, 2, 3}

		blockingQueue := queue.NewBlocking(elems)

		_ = blockingQueue.Clear()

		elem := make(chan int)

		go func() {
			elem <- blockingQueue.GetWait()
		}()

		blockingQueue.OfferWait(4)

		if e := <-elem; e != 4 {
			t.Fatalf("expected elem to be %d, got %d", 4, e)
		}
	})

	t.Run("WithCapacity", func(t *testing.T) {
		t.Parallel()

		elems := []int{1, 2, 3}

		blockingQueue := queue.NewBlocking(
			elems,
			queue.WithCapacity(len(elems)),
		)

		added := make(chan struct{})

		go func() {
			defer close(added)

			blockingQueue.OfferWait(4)
		}()

		select {
		case <-added:
			t.Fatal("received unexpected signal")
		case <-time.After(time.Millisecond):
		}

		for range elems {
			blockingQueue.GetWait()
		}

		if e := blockingQueue.GetWait(); e != 4 {
			t.Fatalf("expected elem to be %d, got %d", 4, e)
		}
	})
}

func testBlockingOffer(t *testing.T) {
	t.Parallel()

	t.Run("NoCapacity", func(t *testing.T) {
		t.Parallel()

		elems := []int{1, 2, 3}

		blockingQueue := queue.NewBlocking[int](elems)

		_ = blockingQueue.Clear()

		elem := make(chan int)

		go func() {
			elem <- blockingQueue.GetWait()
		}()

		if err := blockingQueue.Offer(4); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if e := <-elem; e != 4 {
			t.Fatalf("expected elem to be %d, got %d", 4, e)
		}
	})

	t.Run("WithCapacity", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			elems := []int{1, 2, 3}

			blockingQueue := queue.NewBlocking(
				elems,
				queue.WithCapacity(len(elems)),
			)

			if _, err := blockingQueue.Get(); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if err := blockingQueue.Offer(4); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if e := blockingQueue.GetWait(); e != 2 {
				t.Fatalf("expected elem to be %d, got %d", 2, e)
			}
		})

		t.Run("ErrQueueIsFull", func(t *testing.T) {
			t.Parallel()

			elems := []int{1, 2, 3}

			blockingQueue := queue.NewBlocking(
				elems,
				queue.WithCapacity(len(elems)),
			)

			if err := blockingQueue.Offer(4); !errors.Is(err, queue.ErrQueueIsFull) {
				t.Fatalf("expected error to be %v, got %v", queue.ErrQueueIsFull, err)
			}
		})
	})
}

func testBlockingPeek(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		elems := []int{1, 2, 3}

		blockingQueue := queue.NewBlocking(elems)

		elem, err := blockingQueue.Peek()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if elem != 1 {
			t.Fatalf("expected elem to be %d, got %d", 1, elem)
		}
	})

	t.Run("ErrNoElementsAvailable", func(t *testing.T) {
		t.Parallel()

		blockingQueue := queue.NewBlocking([]int{})

		if _, err := blockingQueue.Peek(); !errors.Is(err, queue.ErrNoElementsAvailable) {
			t.Fatalf("expected error to be %v, got %v", queue.ErrNoElementsAvailable, err)
		}
	})
}

func testBlockingPeekWait(t *testing.T) {
	t.Parallel()

	elems := []int{1, 2, 3}

	blockingQueue := queue.NewBlocking(elems)

	_ = blockingQueue.Clear()

	elem := make(chan int)

	go func() {
		elem <- blockingQueue.PeekWait()
	}()

	time.Sleep(time.Millisecond)

	blockingQueue.OfferWait(4)

	if e := <-elem; e != 4 {
		t.Fatalf("expected elem to be %d, got %d", 4, e)
	}

	if e := blockingQueue.GetWait(); e != 4 {
		t.Fatalf("expected elem to be %d, got %d", 4, e)
	}
}

func testBlockingGet(t *testing.T) {
	t.Parallel()

	elems := []int{1, 2, 3}

	t.Run("ErrNoElementsAvailable", func(t *testing.T) {
		t.Parallel()

		blockingQueue := queue.NewBlocking(elems)

		for range elems {
			blockingQueue.GetWait()
		}

		if _, err := blockingQueue.Get(); !errors.Is(err, queue.ErrNoElementsAvailable) {
			t.Fatalf("expected error to be %v, got %v", queue.ErrNoElementsAvailable, err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		blockingQueue := queue.NewBlocking(elems)

		elem, err := blockingQueue.Get()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if elem != 1 {
			t.Fatalf("expected elem to be %d, got %d", 1, elem)
		}
	})
}

func testBlockingWithCapacity(t *testing.T) {
	t.Parallel()

	elems := []int{1, 2, 3}
	capacity := 2

	blocking := queue.NewBlocking(elems, queue.WithCapacity(capacity))

	if blocking.Size() != capacity {
		t.Fatalf("expected size to be %d, got %d", capacity, blocking.Size())
	}

	if e := blocking.GetWait(); e != 1 {
		t.Fatalf("expected elem to be %d, got %d", 1, e)
	}

	if e := blocking.GetWait(); e != 2 {
		t.Fatalf("expected elem to be %d, got %d", 2, e)
	}

	elem := make(chan int)

	go func() {
		elem <- blocking.GetWait()
	}()

	select {
	case e := <-elem:
		t.Fatalf("received unexepected elem: %d", e)
	case <-time.After(time.Microsecond):
	}

	blocking.OfferWait(4)

	if e := <-elem; e != 4 {
		t.Fatalf("expected elem to be %d, got %d", 4, e)
	}
}

func testBlockingCondWaitWithCapacity(t *testing.T) {
	t.Parallel()

	t.Run("OfferWait", testBlockingCondWaitOfferWait)
	t.Run("GetWait", testBlockingCondWaitGetWait)
	t.Run("PeekWait", testBlockingCondWaitPeekWait)
}

func testBlockingCondWaitOfferWait(t *testing.T) {
	t.Parallel()

	elems := []int{1, 2, 3}
	initialSize := len(elems)

	blockingQueue := queue.NewBlocking(
		elems,
		queue.WithCapacity(initialSize),
	)

	added := make(chan struct{}, initialSize+1)

	for i := 1; i <= initialSize+1; i++ {
		go func(i int) {
			blockingQueue.OfferWait(i)

			added <- struct{}{}
		}(i)
	}

	time.Sleep(time.Millisecond)

	_ = blockingQueue.Clear()

	// one groutine block, and three are added
	for i := 1; i <= initialSize; i++ {
		<-added
	}

	time.Sleep(time.Millisecond)

	if blockingQueue.Size() != initialSize {
		t.Fatalf("expected size to be %d, got %d", initialSize, blockingQueue.Size())
	}

	_ = blockingQueue.GetWait()

	time.Sleep(time.Millisecond)

	if blockingQueue.Size() != initialSize {
		t.Fatalf("expected size to be %d, got %d", initialSize, blockingQueue.Size())
	}
}

func testBlockingCondWaitGetWait(t *testing.T) {
	t.Parallel()

	elems := []int{1, 2, 3}
	initialSize := len(elems)

	blockingQueue := queue.NewBlocking(
		elems,
		queue.WithCapacity(initialSize),
	)

	for i := 1; i <= initialSize; i++ {
		_ = blockingQueue.GetWait()
	}

	if blockingQueue.Size() != 0 {
		t.Fatalf("expected size to be %d, got %d", 0, blockingQueue.Size())
	}

	retrievedElem := make(chan int, initialSize+1)

	for i := 1; i <= initialSize+1; i++ {
		go func() {
			retrievedElem <- blockingQueue.GetWait()
		}()
	}

	time.Sleep(time.Millisecond)
	blockingQueue.Reset()

	// one groutine block, and three are retrieved
	for i := 1; i <= initialSize; i++ {
		<-retrievedElem
	}

	if blockingQueue.Size() != 0 {
		t.Fatalf("expected size to be %d, got %d", initialSize, blockingQueue.Size())
	}

	blockingQueue.OfferWait(4)

	if e := <-retrievedElem; e != 4 {
		t.Fatalf("expected elem to be %d, got %d", 4, e)
	}
}

func testBlockingCondWaitPeekWait(t *testing.T) {
	t.Parallel()

	elems := []int{1}
	initialSize := len(elems)

	blockingQueue := queue.NewBlocking(
		elems,
		queue.WithCapacity(initialSize),
	)

	for i := 1; i <= initialSize; i++ {
		_ = blockingQueue.GetWait()
	}

	if blockingQueue.Size() != 0 {
		t.Fatalf("expected size to be %d, got %d", 0, blockingQueue.Size())
	}

	getCh := make(chan int, 1)

	go func() {
		getCh <- blockingQueue.GetWait()
	}()

	peekCh := make(chan int, 1)

	go func() {
		peekCh <- blockingQueue.PeekWait()
	}()

	time.Sleep(time.Millisecond)
	blockingQueue.Reset()
	// If GetWait is called before PeekWait, PeekWait will block
	// If PeekWait is called before GetWait, PeekWait will not block
	select {
	case <-getCh:
		select {
		case <-peekCh:
		case <-time.After(time.Millisecond):
			t.Log("GetWait is called before PeekWait")
		}
	case <-peekCh:
		select {
		case <-getCh:
			t.Log("PeekWait is called before GetWait")
		case <-time.After(time.Millisecond):
			t.Fatal("expected GetWait to not block")
		}
	case <-time.After(time.Millisecond):
		t.Fatal("expected GetWait or PeekWait not block")
	}

	if blockingQueue.Size() != 0 {
		t.Fatalf("expected size to be %d, got %d", 0, blockingQueue.Size())
	}
}

func testBlockingMarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("HasElements", func(t *testing.T) {
		t.Parallel()

		elems := []int{3, 2, 1}

		q := queue.NewBlocking(elems)

		marshaled, err := json.Marshal(q)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedMarshaled := []byte(`[3,2,1]`)
		if !bytes.Equal(expectedMarshaled, marshaled) {
			t.Fatalf("expected marshaled to be %s, got %s", expectedMarshaled, marshaled)
		}
	})

	t.Run("FailMarshal", func(t *testing.T) {
		t.Parallel()

		q := queue.NewBlocking([]failMarshal{{}})

		marshaled, err := json.Marshal(q)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if marshaled != nil {
			t.Fatalf("expected marshaled to be nil, got %s", marshaled)
		}
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()

		q := queue.NewBlocking[int](nil)

		marshaled, err := json.Marshal(q)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedMarshaled := []byte(`[]`)
		if !bytes.Equal(expectedMarshaled, marshaled) {
			t.Fatalf("expected marshaled to be %s, got %s", expectedMarshaled, marshaled)
		}
	})
}

// failMarshal is a helper to fail the json marshalling of the queues.
type failMarshal struct{}

var errFailMarshal = errors.New("intentional marshal error")

func (failMarshal) MarshalJSON() ([]byte, error) {
	return nil, errFailMarshal
}

func testResetOnMultipleRoutinesFunc[T comparable](
	ids []T,
	totalRoutines int,
) func(t *testing.T) {
	// nolint: thelper // not a test helper
	return func(t *testing.T) {
		blockingQueue := queue.NewBlocking(ids)

		var wg sync.WaitGroup

		wg.Add(totalRoutines)

		retrievedID := make(chan T, len(ids))

		// we start X number of goroutines where X is the total number
		// of goroutines to be executed during this test.
		for routineIdx := 0; routineIdx < totalRoutines; routineIdx++ {
			go func(k int) {
				defer wg.Done()

				t.Logf("start routine %d", k)

				var id T

				defer func() {
					t.Logf("done routine %d, id %v", k, id)
				}()

				retrievedID <- blockingQueue.GetWait()
			}(routineIdx)
		}

		routineCounter := 0

		for range retrievedID {
			routineCounter++

			t.Logf(
				"routine counter: %d, refill: %t",
				routineCounter,
				routineCounter%len(ids) == 0,
			)

			if routineCounter == totalRoutines {
				break
			}

			if routineCounter%len(ids) == 0 {
				blockingQueue.Reset()
			}
		}

		wg.Wait()
	}
}

func BenchmarkBlockingQueue(b *testing.B) {
	b.Run("Peek", func(b *testing.B) {
		blockingQueue := queue.NewBlocking([]int{1})

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_, _ = blockingQueue.Peek()
		}
	})

	b.Run("Get_Offer", func(b *testing.B) {
		blockingQueue := queue.NewBlocking([]int{1})

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_, _ = blockingQueue.Get()

			_ = blockingQueue.Offer(1)
		}
	})

	b.Run("Offer", func(b *testing.B) {
		blockingQueue := queue.NewBlocking[int](nil)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_ = blockingQueue.Offer(i)
		}
	})
}
