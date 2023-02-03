package queue_test

import (
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/adrianbrad/queue"
	"github.com/matryer/is"
)

func TestBlocking(t *testing.T) {
	t.Parallel()

	t.Run("Consistency", func(t *testing.T) {
		t.Parallel()

		i := is.New(t)

		t.Run("SequentialIteration", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			elems := []int{1, 2, 3}

			blockingQueue := queue.NewBlocking(elems)

			for j := range elems {
				id := blockingQueue.GetWait()

				i.Equal(elems[j], id)
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

			i.Equal(ids, result)
		})

		t.Run("PeekWaitAndPushWaiting", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

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
				fmt.Println("peek done")
				i.Equal(elems[0], elem)
			}()

			go func() {
				defer wg.Done()
				<-peekDone

				elem := blockingQueue.GetWait()
				i.Equal(elems[0], elem)
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

	t.Run("Reset", func(t *testing.T) {
		t.Parallel()

		t.Run("WithCapacity", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			elems := []int{1, 2, 3}

			blockingQueue := queue.NewBlocking(
				elems,
				queue.WithCapacity(len(elems)+1),
			)

			i.Equal(3, blockingQueue.Size())

			blockingQueue.OfferWait(4)

			i.Equal(4, blockingQueue.Size())

			blockingQueue.Reset()

			i.Equal(3, blockingQueue.Size())

			size := blockingQueue.Size()

			// empty the queue
			for j := 0; j < size; j++ {
				i.Equal(elems[j], blockingQueue.GetWait())
			}

			elem := make(chan int)

			go func() {
				elem <- blockingQueue.GetWait()
			}()

			blockingQueue.OfferWait(5)

			i.Equal(5, <-elem)
		})
	})

	t.Run("OfferWait", func(t *testing.T) {
		t.Parallel()

		t.Run("NoCapacity", func(t *testing.T) {
			i := is.New(t)

			t.Parallel()

			elems := []int{1, 2, 3}

			blockingQueue := queue.NewBlocking(elems)

			drainQueue[int](blockingQueue)

			elem := make(chan int)

			go func() {
				elem <- blockingQueue.GetWait()
			}()

			blockingQueue.OfferWait(4)

			i.Equal(4, <-elem)
		})

		t.Run("WithCapacity", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

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

			i.Equal(4, blockingQueue.GetWait())
		})
	})

	t.Run("Offer", func(t *testing.T) {
		t.Parallel()

		t.Run("NoCapacity", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			elems := []int{1, 2, 3}

			blockingQueue := queue.NewBlocking[int](elems)

			drainQueue[int](blockingQueue)

			elem := make(chan int)

			go func() {
				elem <- blockingQueue.GetWait()
			}()

			err := blockingQueue.Offer(4)
			i.NoErr(err)

			i.Equal(4, <-elem)
		})

		t.Run("WithCapacity", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				t.Parallel()

				i := is.New(t)

				elems := []int{1, 2, 3}

				blockingQueue := queue.NewBlocking(
					elems,
					queue.WithCapacity(len(elems)),
				)

				_, err := blockingQueue.Get()
				i.NoErr(err)

				err = blockingQueue.Offer(4)
				i.NoErr(err)

				i.Equal(2, blockingQueue.GetWait())
			})

			t.Run("ErrQueueIsFull", func(t *testing.T) {
				t.Parallel()

				i := is.New(t)

				elems := []int{1, 2, 3}

				blockingQueue := queue.NewBlocking(
					elems,
					queue.WithCapacity(len(elems)),
				)

				err := blockingQueue.Offer(4)
				i.Equal(queue.ErrQueueIsFull, err)
			})
		})
	})

	t.Run("Peek", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			elems := []int{1, 2, 3}

			blockingQueue := queue.NewBlocking(elems)

			elem, err := blockingQueue.Peek()
			i.NoErr(err)

			i.Equal(1, elem)
		})

		t.Run("ErrNoElementsAvailable", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			blockingQueue := queue.NewBlocking([]int{})

			_, err := blockingQueue.Peek()
			i.Equal(queue.ErrNoElementsAvailable, err)
		})
	})

	t.Run("PeekWait", func(t *testing.T) {
		t.Parallel()

		i := is.New(t)

		elems := []int{1, 2, 3}

		blockingQueue := queue.NewBlocking(elems)

		drainQueue[int](blockingQueue)

		elem := make(chan int)

		go func() {
			elem <- blockingQueue.PeekWait()
		}()

		time.Sleep(time.Millisecond)

		blockingQueue.OfferWait(4)

		i.Equal(4, <-elem)
		i.Equal(4, blockingQueue.GetWait())
	})

	t.Run("Get", func(t *testing.T) {
		t.Parallel()

		i := is.New(t)

		elems := []int{1, 2, 3}

		t.Run("ErrNoElementsAvailable", func(t *testing.T) {
			t.Parallel()

			blockingQueue := queue.NewBlocking(elems)

			for range elems {
				blockingQueue.GetWait()
			}

			_, err := blockingQueue.Get()
			i.Equal(queue.ErrNoElementsAvailable, err)
		})

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			blockingQueue := queue.NewBlocking(elems)

			elem, err := blockingQueue.Get()
			i.NoErr(err)

			i.Equal(1, elem)
		})
	})

	t.Run("WithCapacity", func(t *testing.T) {
		t.Parallel()

		i := is.New(t)

		elems := []int{1, 2, 3}
		capacity := 2

		blocking := queue.NewBlocking(elems, queue.WithCapacity(capacity))

		i.Equal(2, blocking.Size())

		i.Equal(1, blocking.GetWait())
		i.Equal(2, blocking.GetWait())

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

		i.Equal(4, <-elem)
	})
}

func testResetOnMultipleRoutinesFunc[T any](
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
