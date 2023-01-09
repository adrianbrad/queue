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

	ids := []string{"0", "1", "2"}

	t.Run("Consistency", func(t *testing.T) {
		i := is.New(t)

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

	t.Run("SequentialIteration", func(t *testing.T) {
		t.Parallel()

		i := is.New(t)

		blockingQueue := queue.NewBlocking(ids)

		for j := range ids {
			id := blockingQueue.Take()

			i.Equal(ids[j], id)
		}
	})

	t.Run("Reset", func(t *testing.T) {
		t.Parallel()

		t.Run("CancelContext", func(t *testing.T) {
			t.Parallel()

			t.Run("BeforeRefill", func(t *testing.T) {
				t.Parallel()

				blockingQueue := queue.NewBlocking(ids)

				blockingQueue.Take()

				done := make(chan struct{})

				go func() {
					blockingQueue.Reset()
					close(done)
				}()

				select {
				case <-done:
					return

				case <-time.After(time.Second):
					t.Error("refill was supposed to return")
				}
			})
		})

		t.Run("SequentialRefill", func(t *testing.T) {
			t.Parallel()

			const noRoutines = 100

			for i := 1; i <= noRoutines; i++ {
				i := i

				t.Run(
					fmt.Sprintf("%dRoutinesWaiting", i),
					func(t *testing.T) {
						t.Parallel()

						testRefillOnMultipleRoutinesFunc[string](ids, i)(t)
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

func TestBlocking_Push(t *testing.T) {
	i := is.New(t)

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
}

func TestBlocking_Peek(t *testing.T) {
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
	i := is.New(t)

	elems := []int{1, 2, 3}

	blockingQueue := queue.NewBlocking(elems)

	for range elems {
		blockingQueue.Take()
	}

	_, err := blockingQueue.Get()
	i.Equal(queue.ErrNoElementsAvailable, err)
}
