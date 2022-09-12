package queue_test

import (
	"context"
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
	ctx := context.Background()

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

		go blockingQueue.Refill(ctx)

		for i := 0; i < lenElements; i++ {
			go func() {
				elem := blockingQueue.Take(ctx)

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
			id := blockingQueue.Take(ctx)

			i.Equal(ids[j], id)
		}
	})

	t.Run("CancelContext", func(t *testing.T) {
		t.Parallel()

		i := is.New(t)

		blockingQueue := queue.NewBlocking(ids)

		for range ids {
			blockingQueue.Take(ctx)
		}

		ctx, cancelCtx := context.WithCancel(ctx)
		cancelCtx()

		e := blockingQueue.Take(ctx)

		i.Equal("", e)
	})

	t.Run("Refill", func(t *testing.T) {
		t.Parallel()

		t.Run("CancelContext", func(t *testing.T) {
			t.Parallel()

			t.Run("BeforeRefill", func(t *testing.T) {
				t.Parallel()

				blockingQueue := queue.NewBlocking(ids)

				refillCtx, cancelRefillCtx := context.WithCancel(ctx)

				blockingQueue.Take(ctx)

				done := make(chan struct{})

				cancelRefillCtx()

				go func() {
					blockingQueue.Refill(refillCtx)
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

						testRefillOnMultipleRoutinesFunc[string](ctx, ids, i)(t)
					},
				)
			}
		})
	})
}

func testRefillOnMultipleRoutinesFunc[T any](
	ctx context.Context,
	ids []T,
	totalRoutines int,
) func(t *testing.T) {
	// nolint: thelper // not a test helper
	return func(t *testing.T) {
		blockingQueue := queue.NewBlocking(ids)

		for range ids {
			blockingQueue.Take(ctx)
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

				id = blockingQueue.Take(ctx)

				retrievedID <- id
			}(routineIdx)
		}

		time.Sleep(time.Millisecond)

		t.Log("refill")

		blockingQueue.Refill(ctx)

		counter := 0

		for range retrievedID {
			counter++

			t.Logf(
				"counter: %d, refill: %t",
				counter,
				counter%len(ids) == 0,
			)

			if counter == totalRoutines {
				break
			}

			if counter%len(ids) == 0 {
				blockingQueue.Refill(ctx)
			}
		}

		wg.Wait()
	}
}
