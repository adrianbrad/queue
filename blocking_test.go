package queue_test

import (
	"context"
	"fmt"
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

		ctx, cancelCtx := context.WithTimeout(ctx, time.Millisecond)
		defer cancelCtx()

		e := blockingQueue.Take(ctx)

		i.Equal("", e)
	})

	t.Run("Reset", func(t *testing.T) {
		t.Parallel()

		t.Run("SequentialReset", func(t *testing.T) {
			t.Parallel()

			const noRoutines = 30

			for i := 1; i <= noRoutines; i++ {
				i := i

				t.Run(
					fmt.Sprintf("%dRoutinesWaiting", i),
					func(t *testing.T) {
						t.Parallel()

						testResetOnMultipleRoutinesFunc[string](ctx, ids, i)(t)
					},
				)
			}
		})
	})
}

func testResetOnMultipleRoutinesFunc[T any](
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

		t.Log("reset")

		blockingQueue.Reset()

		counter := 0

		for range retrievedID {
			counter++

			t.Logf(
				"counter: %d, reset: %t",
				counter,
				counter%len(ids) == 0,
			)

			if counter == totalRoutines {
				break
			}

			if counter%len(ids) == 0 {
				blockingQueue.Reset()
			}
		}

		wg.Wait()
	}
}
