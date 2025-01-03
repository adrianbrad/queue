package queue_test

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/adrianbrad/queue"
)

func TestPriority(t *testing.T) {
	t.Parallel()

	lessAscending := func(elem, elemAfter int) bool {
		return elem < elemAfter
	}

	lessInt := func(elem, elemAfter int) bool {
		return elem < elemAfter
	}

	t.Run("NilLessFunc", func(t *testing.T) {
		defer func() {
			p := recover()
			if p != "nil less func" {
				t.Fatalf("expected panic to be 'nil less func', got %v", p)
			}
		}()

		queue.NewPriority[any](nil, nil)
	})

	t.Run("CapacityLesserThanLenElems", func(t *testing.T) {
		t.Parallel()

		elems := []int{4, 1, 2}

		priorityQueue := queue.NewPriority(elems, lessInt, queue.WithCapacity(2))

		size := priorityQueue.Size()

		if priorityQueue.Size() != 2 {
			t.Fatalf("expected size to be 2, got %d with elements: %v", size, priorityQueue.Clear())
		}

		elems = priorityQueue.Clear()
		expectedElems := []int{1, 2}

		if !reflect.DeepEqual([]int{1, 2}, elems) {
			t.Fatalf("expected elements to be %v, got %v", expectedElems, elems)
		}
	})

	t.Run("Offer", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			elems := []int{4, 1, 2}

			priorityQueue := queue.NewPriority(elems, lessAscending)

			if err := priorityQueue.Offer(5); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			size := priorityQueue.Size()

			if size != 4 {
				t.Fatalf("expected size to be 4, got %d", size)
			}

			queueElems := priorityQueue.Clear()
			expectedElems := []int{1, 2, 4, 5}

			if !reflect.DeepEqual(expectedElems, queueElems) {
				t.Fatalf("expected elements to be %v, got %v", expectedElems, queueElems)
			}

			newElems := make([]int, 10)

			for j := 19; j >= 10; j-- {
				newElems[j%10] = j

				if err := priorityQueue.Offer(j); err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}

			queueElems = priorityQueue.Clear()
			if !reflect.DeepEqual(newElems, queueElems) {
				t.Fatalf("expected elements to be %v, got %v", newElems, queueElems)
			}
		})

		t.Run("ErrQueueIsFull", func(t *testing.T) {
			t.Parallel()

			elems := []int{1}

			priorityQueue := queue.NewPriority(
				elems, lessAscending,
				queue.WithCapacity(1),
			)

			if err := priorityQueue.Offer(2); !errors.Is(err, queue.ErrQueueIsFull) {
				t.Fatalf("expected error to be %v, got %v", queue.ErrQueueIsFull, err)
			}
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			elems := []int{4, 1, 2}

			priorityQueue := queue.NewPriority(elems, lessInt)

			elem, err := priorityQueue.Get()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if elem != 1 {
				t.Fatalf("expected elem to be 1, got %d", elem)
			}

			size := priorityQueue.Size()

			if size != 2 {
				t.Fatalf("expected size to be 2, got %d", size)
			}
		})

		t.Run("ErrNoElementsAvailable", func(t *testing.T) {
			t.Parallel()

			priorityQueue := queue.NewPriority([]int{}, func(_, _ int) bool { return false })

			if _, err := priorityQueue.Get(); !errors.Is(err, queue.ErrNoElementsAvailable) {
				t.Fatalf("expected error to be %v, got %v", queue.ErrNoElementsAvailable, err)
			}
		})
	})

	t.Run("Clear", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			elems := []int{1, 2, 3}

			priorityQueue := queue.NewPriority(elems, lessAscending)

			queueElems := priorityQueue.Clear()

			if !reflect.DeepEqual(elems, queueElems) {
				t.Fatalf("expected elements to be %v, got %v", elems, queueElems)
			}
		})

		t.Run("Empty", func(t *testing.T) {
			t.Parallel()

			priorityQueue := queue.NewPriority([]int{}, lessAscending)

			queueElems := priorityQueue.Clear()

			if len(queueElems) != 0 {
				t.Fatalf("expected elements to be empty, got %v", queueElems)
			}
		})
	})

	t.Run("Contains", func(t *testing.T) {
		t.Parallel()

		t.Run("True", func(t *testing.T) {
			t.Parallel()

			elems := []int{1, 2, 3}

			priorityQueue := queue.NewPriority(elems, lessAscending)

			if !priorityQueue.Contains(2) {
				t.Fatalf("expected queue to contain 2")
			}
		})

		t.Run("False", func(t *testing.T) {
			t.Parallel()

			elems := []int{1, 2, 3}

			priorityQueue := queue.NewPriority(elems, lessAscending)

			if priorityQueue.Contains(4) {
				t.Fatalf("expected queue to not contain 4")
			}
		})
	})

	t.Run("Iterator", func(t *testing.T) {
		t.Parallel()

		elems := []int{1, 2, 3}

		priorityQueue := queue.NewPriority(elems, lessAscending)

		iterCh := priorityQueue.Iterator()

		if !priorityQueue.IsEmpty() {
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

	t.Run("IsEmpty", func(t *testing.T) {
		t.Parallel()

		t.Run("True", func(t *testing.T) {
			t.Parallel()

			priorityQueue := queue.NewPriority([]int{}, lessAscending)

			if !priorityQueue.IsEmpty() {
				t.Fatalf("expected queue to be empty")
			}
		})

		t.Run("False", func(t *testing.T) {
			t.Parallel()

			priorityQueue := queue.NewPriority([]int{1}, lessAscending)

			if priorityQueue.IsEmpty() {
				t.Fatalf("expected queue to not be empty")
			}
		})
	})

	t.Run("Peek", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			elems := []int{4, 1, 2}

			priorityQueue := queue.NewPriority(elems, lessAscending)

			elem, err := priorityQueue.Peek()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			size := priorityQueue.Size()

			if size != 3 {
				t.Fatalf("expected size to be 3, got %d", size)
			}

			if elem != 1 {
				t.Fatalf("expected elem to be 1, got %d", elem)
			}

			elem, err = priorityQueue.Get()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if elem != 1 {
				t.Fatalf("expected elem to be 1, got %d", elem)
			}

			elem, err = priorityQueue.Peek()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			size = priorityQueue.Size()

			if size != 2 {
				t.Fatalf("expected size to be 2, got %d", size)
			}

			if elem != 2 {
				t.Fatalf("expected elem to be 2, got %d", elem)
			}

			elem, err = priorityQueue.Get()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if elem != 2 {
				t.Fatalf("expected elem to be 2, got %d", elem)
			}
		})

		t.Run("ErrNoElementsAvailable", func(t *testing.T) {
			t.Parallel()

			priorityQueue := queue.NewPriority([]int{}, func(_, _ int) bool { return false })

			if _, err := priorityQueue.Peek(); !errors.Is(err, queue.ErrNoElementsAvailable) {
				t.Fatalf("expected error to be %v, got %v", queue.ErrNoElementsAvailable, err)
			}
		})
	})

	t.Run("Reset", func(t *testing.T) {
		t.Run("SizeGreaterThanInitialElems", func(t *testing.T) {
			t.Parallel()

			elems := []int{1}

			priorityQueue := queue.NewPriority(elems, lessAscending)

			err := priorityQueue.Offer(2)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			priorityQueue.Reset()

			if priorityQueue.Size() != 1 {
				t.Fatalf("expected size to be 1, got %d", priorityQueue.Size())
			}
		})

		t.Run("SizeLesserThanInitialElems", func(t *testing.T) {
			t.Parallel()

			elems := []int{1, 2}

			priorityQueue := queue.NewPriority(elems, lessAscending)

			if _, err := priorityQueue.Get(); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			priorityQueue.Reset()

			if priorityQueue.Size() != 2 {
				t.Fatalf("expected size to be 2, got %d", priorityQueue.Size())
			}
		})
	})
}

func FuzzPriority(f *testing.F) {
	testcases := [][]byte{{2, 10, 8, 4}, {11, 9, 7}, {8, 24, 255}}
	for _, tc := range testcases {
		f.Add(tc)
	}

	lessFunc := func(elem, elemAfter byte) bool {
		return elem < elemAfter
	}

	f.Fuzz(func(t *testing.T, orig []byte) {
		sort.Slice(orig, func(i, j int) bool {
			return orig[i] < orig[j]
		})

		priorityQueue := queue.NewPriority(nil, lessFunc)

		for _, v := range orig {
			err := priorityQueue.Offer(v)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		}

		for _, v := range orig {
			peekedVal, err := priorityQueue.Peek()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if v != peekedVal {
				t.Fatalf("expected peeked value to be %d, got %d", v, peekedVal)
			}

			getVal, err := priorityQueue.Get()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if peekedVal != getVal {
				t.Fatalf("expected peeked value to be %d, got %d", peekedVal, getVal)
			}
		}
	})
}

func BenchmarkPriorityQueue(b *testing.B) {
	b.Run("Peek", func(b *testing.B) {
		priorityQueue := queue.NewPriority([]int{1}, func(elem, otherElem int) bool {
			return elem < otherElem
		})

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_, _ = priorityQueue.Peek()
		}
	})

	b.Run("Get_Offer", func(b *testing.B) {
		priorityQueue := queue.NewPriority([]int{1}, func(elem, otherElem int) bool {
			return elem < otherElem
		})

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_, _ = priorityQueue.Get()

			_ = priorityQueue.Offer(1)
		}
	})

	b.Run("Offer", func(b *testing.B) {
		priorityQueue := queue.NewPriority[int](nil, func(elem, otherElem int) bool {
			return elem < otherElem
		})

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_ = priorityQueue.Offer(i)
		}
	})
}
