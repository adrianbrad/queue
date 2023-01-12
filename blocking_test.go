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

		t.Run("100 concurrent goroutines reading", func(t *testing.T) {
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
					elem := blockingQueue.Take()

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

		t.Run("Peek and Push Waiting", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			elems := []int{1}

			blockingQueue := queue.NewBlocking(elems)

			_ = blockingQueue.Take()

			var wg sync.WaitGroup

			wg.Add(2)

			peekDone := make(chan struct{})

			go func() {
				defer wg.Done()
				defer close(peekDone)

				elem := blockingQueue.Peek()
				i.Equal(elems[0], elem)
			}()

			go func() {
				defer wg.Done()
				<-peekDone

				elem := blockingQueue.Take()
				i.Equal(elems[0], elem)
			}()

			time.Sleep(time.Microsecond)

			blockingQueue.Reset()

			wg.Wait()
		})
	})

	t.Run("SequentialIteration", func(t *testing.T) {
		t.Parallel()

		i := is.New(t)

		elems := []int{1, 2, 3}

		blockingQueue := queue.NewBlocking(elems)

		for j := range elems {
			id := blockingQueue.Take()

			i.Equal(elems[j], id)
		}
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

			blockingQueue.Put(4)

			i.Equal(4, blockingQueue.Size())

			blockingQueue.Reset()

			i.Equal(3, blockingQueue.Size())

			size := blockingQueue.Size()

			for j := 0; j < size; j++ {
				i.Equal(elems[j], blockingQueue.Take())
			}

			elem := make(chan int)

			go func() {
				elem <- blockingQueue.Take()
			}()

			time.Sleep(time.Millisecond)

			blockingQueue.Put(5)

			i.Equal(5, <-elem)
		})

		t.Run("SequentialRefill", func(t *testing.T) {
			t.Parallel()

			elems := []int{1, 2, 3}

			const noRoutines = 100

			for i := 1; i <= noRoutines; i++ {
				i := i

				t.Run(
					fmt.Sprintf("%dRoutinesWaiting", i),
					func(t *testing.T) {
						t.Parallel()

						testRefillOnMultipleRoutinesFunc[int](elems, i)(t)
					},
				)
			}
		})
	})
}

func testRefillOnMultipleRoutinesFunc[T any](
	ids []T,
	totalRoutines int,
) func(t *testing.T) {
	// nolint: thelper // not a test helper
	return func(t *testing.T) {
		blockingQueue := queue.NewBlocking(ids)

		// empty the queue
		for range ids {
			blockingQueue.Take()
		}

		var wg sync.WaitGroup

		wg.Add(totalRoutines)

		retrievedID := make(chan T, len(ids))

		for routineIdx := 0; routineIdx < totalRoutines; routineIdx++ {
			go func(k int) {
				defer wg.Done()

				t.Logf("start routine %d", k)

				var id T

				defer func() {
					t.Logf("done routine %d, id %v", k, id)
				}()

				retrievedID <- blockingQueue.Take()
			}(routineIdx)
		}

		time.Sleep(time.Millisecond)

		t.Log("refill")

		// refill with 3 elems
		blockingQueue.Reset()

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

func TestBlocking_Put(t *testing.T) {
	t.Parallel()

	t.Run("NoCapacity", func(t *testing.T) {
		i := is.New(t)

		t.Parallel()

		elems := []int{1, 2, 3}

		blockingQueue := queue.NewBlocking(elems)

		for range elems {
			blockingQueue.Take()
		}

		elem := make(chan int)

		go func() {
			elem <- blockingQueue.Take()
		}()

		time.Sleep(time.Millisecond)

		blockingQueue.Put(4)

		i.Equal(4, <-elem)
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

			blockingQueue.Put(4)
		}()

		select {
		case <-added:
			t.Fatalf("received unexpected signal")
		case <-time.After(time.Millisecond):
		}

		for range elems {
			blockingQueue.Take()
		}

		<-added
	})
}

func TestBlocking_Peek(t *testing.T) {
	t.Parallel()

	i := is.New(t)

	elems := []int{1, 2, 3}

	blockingQueue := queue.NewBlocking(elems)

	for range elems {
		blockingQueue.Take()
	}

	elem := make(chan int)

	go func() {
		elem <- blockingQueue.Peek()
	}()

	time.Sleep(time.Millisecond)

	blockingQueue.Put(4)

	i.Equal(4, <-elem)
	i.Equal(4, blockingQueue.Take())
}

func TestBlocking_Get(t *testing.T) {
	t.Parallel()

	i := is.New(t)

	elems := []int{1, 2, 3}

	t.Run("NoElemsAvailable", func(t *testing.T) {
		t.Parallel()

		blockingQueue := queue.NewBlocking(elems)

		for range elems {
			blockingQueue.Take()
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
}

func TestBlocking_Capacity(t *testing.T) {
	t.Parallel()

	i := is.New(t)

	elems := []int{1, 2, 3}
	capacity := 2

	blocking := queue.NewBlocking(elems, queue.WithCapacity(capacity))

	i.Equal(2, blocking.Size())

	i.Equal(1, blocking.Take())
	i.Equal(2, blocking.Take())

	elem := make(chan int)

	go func() {
		elem <- blocking.Take()
	}()

	select {
	case e := <-elem:
		t.Fatalf("received unexepected elem: %d", e)
	case <-time.After(time.Microsecond):
	}

	blocking.Put(4)

	i.Equal(4, <-elem)
}
