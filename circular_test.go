package queue_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/adrianbrad/queue"
)

func TestCircular(t *testing.T) {
	t.Parallel()

	t.Run("CapcaityOptionsOverwritesCapacityParam", func(t *testing.T) {
		t.Parallel()

		circularQueue := queue.NewCircular([]int{}, 1, queue.WithCapacity(2))

		circularQueue.Offer(1)
		circularQueue.Offer(2)

		if circularQueue.Size() != 2 {
			t.Fatalf("expected size to be 2, got %d", circularQueue.Size())
		}
	})

	t.Run("ElemsLenGreaterThanCapacity", func(t *testing.T) {
		t.Parallel()

		circularQueue := queue.NewCircular([]int{1, 2}, 1)

		if circularQueue.Size() != 1 {
			t.Fatalf("expected size to be 1, got %d", circularQueue.Size())
		}
	})

	t.Run("Get", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			elems := []int{4, 1, 2}

			circularQueue := queue.NewCircular(elems, len(elems))

			elem, err := circularQueue.Get()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if elem != 4 {
				t.Fatalf("expected elem to be 4, got %d", elem)
			}

			if circularQueue.Size() != 2 {
				t.Fatalf("expected size to be 2, got %d", circularQueue.Size())
			}
		})

		t.Run("ErrNoElementsAvailable", func(t *testing.T) {
			t.Parallel()

			var elems []int

			circularQueue := queue.NewCircular(elems, 5)

			if _, err := circularQueue.Get(); !errors.Is(err, queue.ErrNoElementsAvailable) {
				t.Fatalf("expected error to be %v, got %v", queue.ErrNoElementsAvailable, err)
			}
		})
	})

	t.Run("Peek", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			elems := []int{4, 1, 2}

			circularQueue := queue.NewCircular(elems, len(elems))

			elem, err := circularQueue.Peek()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if elem != 4 {
				t.Fatalf("expected elem to be 4, got %d", elem)
			}

			if circularQueue.Size() != 3 {
				t.Fatalf("expected size to be 3, got %d", circularQueue.Size())
			}
		})

		t.Run("ErrNoElementsAvailable", func(t *testing.T) {
			t.Parallel()

			var elems []int

			circularQueue := queue.NewCircular(elems, 5)

			if _, err := circularQueue.Peek(); !errors.Is(err, queue.ErrNoElementsAvailable) {
				t.Fatalf("expected error to be %v, got %v", queue.ErrNoElementsAvailable, err)
			}
		})
	})

	t.Run("Offer", func(t *testing.T) {
		t.Parallel()

		t.Run("SuccessEmptyQueue", func(t *testing.T) {
			var elems []int

			circularQueue := queue.NewCircular(elems, 5)

			err := circularQueue.Offer(1)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			err = circularQueue.Offer(2)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if circularQueue.Size() != 2 {
				t.Fatalf("expected size to be 2, got %d", circularQueue.Size())
			}

			queueElems := circularQueue.Clear()
			expectedElems := []int{1, 2}

			if !reflect.DeepEqual(expectedElems, queueElems) {
				t.Fatalf("expected elems to be %v, got %v", expectedElems, queueElems)
			}
		})

		t.Run("SuccessFullQueue", func(t *testing.T) {
			elems := []int{1, 2, 3, 4}

			circularQueue := queue.NewCircular(elems, 4)

			err := circularQueue.Offer(5)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if circularQueue.Size() != 4 {
				t.Fatalf("expected size to be 4, got %d", circularQueue.Size())
			}

			nextElem, err := circularQueue.Peek()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if nextElem != 5 {
				t.Fatalf("expected next elem to be 4, got %d", nextElem)
			}

			err = circularQueue.Offer(6)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			queueElems := circularQueue.Clear()
			expectedElems := []int{5, 6, 3, 4}

			if !reflect.DeepEqual(expectedElems, queueElems) {
				t.Fatalf("expected elems to be %v, got %v", expectedElems, queueElems)
			}
		})
	})

	t.Run("Contains", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			elems := []int{1, 2, 3, 4}

			circularQueue := queue.NewCircular(elems, 4)

			if !circularQueue.Contains(2) {
				t.Fatalf("expected elem to be found")
			}
		})

		t.Run("NotFoundAfterGet", func(t *testing.T) {
			t.Parallel()

			elems := []int{1, 2, 3, 4}

			circularQueue := queue.NewCircular(elems, 4)

			_, err := circularQueue.Get()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if circularQueue.Contains(1) {
				t.Fatalf("expected elem to not be found")
			}
		})

		t.Run("EmptyQueue", func(t *testing.T) {
			circularQueue := queue.NewCircular([]int{}, 1)

			if circularQueue.Contains(1) {
				t.Fatalf("expected elem to not be found")
			}
		})
	})

	t.Run("Clear", func(t *testing.T) {
		t.Parallel()

		elems := []int{1, 2, 3, 4}

		circularQueue := queue.NewCircular(elems, 4)

		_, err := circularQueue.Get()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		queueElems := circularQueue.Clear()
		expectedElems := []int{2, 3, 4}

		if !reflect.DeepEqual(expectedElems, queueElems) {
			t.Fatalf("expected elems to be %v, got %v", expectedElems, queueElems)
		}
	})

	t.Run("IsEmpty", func(t *testing.T) {
		circularQueue := queue.NewCircular([]int{}, 1)

		if !circularQueue.IsEmpty() {
			t.Fatalf("expected queue to be empty")
		}
	})

	t.Run("Reset", func(t *testing.T) {
		elems := []int{1, 2, 3, 4}

		circularQueue := queue.NewCircular(elems, 5)

		err := circularQueue.Offer(5)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		err = circularQueue.Offer(6)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		circularQueue.Reset()

		queueElems := circularQueue.Clear()
		expectedElems := []int{1, 2, 3, 4}

		if !reflect.DeepEqual(expectedElems, queueElems) {
			t.Fatalf("expected elems to be %v, got %v", expectedElems, queueElems)
		}
	})

	t.Run("Iterator", func(t *testing.T) {
		elems := []int{1, 2, 3, 4}

		circularQueue := queue.NewCircular(elems, 5)

		iterCh := circularQueue.Iterator()

		if !circularQueue.IsEmpty() {
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
}

func BenchmarkCircularQueue(b *testing.B) {
	b.Run("Peek", func(b *testing.B) {
		circularQueue := queue.NewCircular([]int{1}, 1)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_, _ = circularQueue.Peek()
		}
	})

	b.Run("Get_Offer", func(b *testing.B) {
		circularQueue := queue.NewCircular([]int{1}, 1)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_, _ = circularQueue.Get()

			_ = circularQueue.Offer(1)
		}
	})

	b.Run("Offer", func(b *testing.B) {
		circularQueue := queue.NewCircular[int](nil, 1)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_ = circularQueue.Offer(i)
		}
	})
}
