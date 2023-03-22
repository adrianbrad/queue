package queue_test

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/adrianbrad/queue/v2"
)

func TestBlocking(t *testing.T) {
	t.Parallel()

	t.Run("Consistency", func(t *testing.T) {
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

				t.Logf("peek done")

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
				i := i

				t.Run(
					fmt.Sprintf("%dRoutinesWaiting", i),
					func(t *testing.T) {
						testResetOnMultipleRoutinesFunc[int](elems, i)(t)
					},
				)
			}
		})
	})

	t.Run("Clear", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			elems := []int{1, 2, 3}

			blockingQueue := queue.NewBlocking(elems)

			queueElems := blockingQueue.Clear()

			if !reflect.DeepEqual(elems, queueElems) {
				t.Fatalf("expected elems to be %v, got %v", elems, queueElems)
			}
		})

		t.Run("Empty", func(t *testing.T) {
			t.Parallel()

			blockingQueue := queue.NewBlocking([]int{})

			queueElems := blockingQueue.Clear()

			if len(queueElems) != 0 {
				t.Fatalf("expected elems to be empty, got %v", queueElems)
			}
		})
	})

	t.Run("Contains", func(t *testing.T) {
		t.Parallel()

		t.Run("True", func(t *testing.T) {
			t.Parallel()

			elems := []int{1, 2, 3}

			blockingQueue := queue.NewBlocking(elems)

			if !blockingQueue.Contains(2) {
				t.Fatalf("expected queue to contain 2")
			}
		})

		t.Run("False", func(t *testing.T) {
			t.Parallel()

			elems := []int{1, 2, 3}

			blockingQueue := queue.NewBlocking(elems)

			if blockingQueue.Contains(4) {
				t.Fatalf("expected queue to not contain 4")
			}
		})
	})

	t.Run("Iterator", func(t *testing.T) {
		t.Parallel()

		elems := []int{1, 2, 3}

		blockingQueue := queue.NewBlocking(elems)

		iterCh := blockingQueue.Iterator()

		if !blockingQueue.IsEmpty() {
			t.Fatalf("expected queue to be empty")
		}

		iterElems := make([]int, 0, len(elems))

		for e := range iterCh {
			iterElems = append(iterElems, e)
		}

		if !reflect.DeepEqual(elems, iterElems) {
			t.Fatalf("expected elems to be %v, got %v", elems, iterElems)
		}
	})

	t.Run("IsEmpty", func(t *testing.T) {
		t.Parallel()

		t.Run("True", func(t *testing.T) {
			t.Parallel()

			blockingQueue := queue.NewBlocking([]int{})

			if !blockingQueue.IsEmpty() {
				t.Fatalf("expected queue to be empty")
			}
		})

		t.Run("False", func(t *testing.T) {
			t.Parallel()

			blockingQueue := queue.NewBlocking([]int{1})

			if blockingQueue.IsEmpty() {
				t.Fatalf("expected queue to not be empty")
			}
		})
	})

	t.Run("Reset", func(t *testing.T) {
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
	})

	t.Run("OfferWait", func(t *testing.T) {
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
				t.Fatalf("received unexpected signal")
			case <-time.After(time.Millisecond):
			}

			for range elems {
				blockingQueue.GetWait()
			}

			if e := blockingQueue.GetWait(); e != 4 {
				t.Fatalf("expected elem to be %d, got %d", 4, e)
			}
		})
	})

	t.Run("Offer", func(t *testing.T) {
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
	})

	t.Run("Peek", func(t *testing.T) {
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
	})

	t.Run("PeekWait", func(t *testing.T) {
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
	})

	t.Run("Get", func(t *testing.T) {
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
	})

	t.Run("WithCapacity", func(t *testing.T) {
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
	})
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
