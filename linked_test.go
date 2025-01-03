package queue_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/adrianbrad/queue"
)

func TestLinked(t *testing.T) {
	t.Parallel()

	t.Run("Get", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			elems := []int{4, 1, 2}

			linkedQueue := queue.NewLinked(elems)

			elem, err := linkedQueue.Get()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if elem != 4 {
				t.Fatalf("expected elem to be 4, got %d", elem)
			}

			if linkedQueue.Size() != 2 {
				t.Fatalf("expected size to be 2, got %d", linkedQueue.Size())
			}
		})

		t.Run("ErrNoElementsAvailable", func(t *testing.T) {
			t.Parallel()

			var elems []int

			linkedQueue := queue.NewLinked(elems)

			if _, err := linkedQueue.Get(); !errors.Is(err, queue.ErrNoElementsAvailable) {
				t.Fatalf("expected error to be %v, got %v", queue.ErrNoElementsAvailable, err)
			}
		})
	})

	t.Run("Peek", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			elems := []int{4, 1, 2}

			linkedQueue := queue.NewLinked(elems)

			elem, err := linkedQueue.Peek()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if elem != 4 {
				t.Fatalf("expected elem to be 4, got %d", elem)
			}

			if linkedQueue.Size() != 3 {
				t.Fatalf("expected size to be 3, got %d", linkedQueue.Size())
			}
		})

		t.Run("ErrNoElementsAvailable", func(t *testing.T) {
			t.Parallel()

			var elems []int

			linkedQueue := queue.NewLinked(elems)

			if _, err := linkedQueue.Peek(); !errors.Is(err, queue.ErrNoElementsAvailable) {
				t.Fatalf("expected error to be %v, got %v", queue.ErrNoElementsAvailable, err)
			}
		})
	})

	t.Run("Offer", func(t *testing.T) {
		t.Parallel()

		t.Run("SuccessEmptyQueue", func(t *testing.T) {
			var elems []int

			linkedQueue := queue.NewLinked(elems)

			err := linkedQueue.Offer(1)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			err = linkedQueue.Offer(2)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if linkedQueue.Size() != 2 {
				t.Fatalf("expected size to be 2, got %d", linkedQueue.Size())
			}

			queueElems := linkedQueue.Clear()
			expectedElems := []int{1, 2}

			if !reflect.DeepEqual(expectedElems, queueElems) {
				t.Fatalf("expected elements to be %v, got %v", expectedElems, queueElems)
			}
		})
	})

	t.Run("Contains", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			elems := []int{1, 2, 3, 4}

			linkedQueue := queue.NewLinked(elems)

			if !linkedQueue.Contains(2) {
				t.Fatalf("expected elem to be found")
			}
		})

		t.Run("NotFoundAfterGet", func(t *testing.T) {
			t.Parallel()

			elems := []int{1, 2, 3, 4}

			linkedQueue := queue.NewLinked(elems)

			_, err := linkedQueue.Get()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if linkedQueue.Contains(1) {
				t.Fatalf("expected elem to not be found")
			}
		})

		t.Run("EmptyQueue", func(t *testing.T) {
			linkedQueue := queue.NewLinked([]int{})

			if linkedQueue.Contains(1) {
				t.Fatalf("expected elem to not be found")
			}
		})
	})

	t.Run("Clear", func(t *testing.T) {
		t.Parallel()

		elems := []int{1, 2, 3, 4}

		linkedQueue := queue.NewLinked(elems)

		_, err := linkedQueue.Get()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		queueElems := linkedQueue.Clear()
		expectedElems := []int{2, 3, 4}

		if !reflect.DeepEqual(expectedElems, queueElems) {
			t.Fatalf("expected elements to be %v, got %v", expectedElems, queueElems)
		}
	})

	t.Run("IsEmpty", func(t *testing.T) {
		linkedQueue := queue.NewLinked([]int{})

		if !linkedQueue.IsEmpty() {
			t.Fatalf("expected queue to be empty")
		}
	})

	t.Run("Reset", func(t *testing.T) {
		elems := []int{1, 2, 3, 4}

		linkedQueue := queue.NewLinked(elems)

		err := linkedQueue.Offer(5)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		err = linkedQueue.Offer(6)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		linkedQueue.Reset()

		queueElems := linkedQueue.Clear()
		expectedElems := []int{1, 2, 3, 4}

		if !reflect.DeepEqual(expectedElems, queueElems) {
			t.Fatalf("expected elements to be %v, got %v", expectedElems, queueElems)
		}
	})

	t.Run("Iterator", func(t *testing.T) {
		elems := []int{1, 2, 3, 4}

		linkedQueue := queue.NewLinked(elems)

		iterCh := linkedQueue.Iterator()

		if !linkedQueue.IsEmpty() {
			t.Fatalf("expected queue to be empty")
		}

		iterElems := make([]int, 0, len(elems))

		for e := range iterCh {
			iterElems = append(iterElems, e)
		}

		if !reflect.DeepEqual(elems, iterElems) {
			t.Fatalf("expected elements to be %v, got %v", elems, iterElems)
		}
	})
}

func BenchmarkLinkedQueue(b *testing.B) {
	b.Run("Peek", func(b *testing.B) {
		linkedQueue := queue.NewLinked([]int{1})

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_, _ = linkedQueue.Peek()
		}
	})

	b.Run("Get_Offer", func(b *testing.B) {
		linkedQueue := queue.NewLinked([]int{1})

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_, _ = linkedQueue.Get()

			_ = linkedQueue.Offer(1)
		}
	})

	b.Run("Offer", func(b *testing.B) {
		linkedQueue := queue.NewLinked[int](nil)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_ = linkedQueue.Offer(i)
		}
	})
}
