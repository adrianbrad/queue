package queue_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/adrianbrad/queue"
)

func TestCircular(t *testing.T) {
	t.Parallel()

	t.Run("CapcaityOptionsOverwritesCapacityParam", testCircularCapacityOptionOverwrites)
	t.Run("ElemsLenGreaterThanCapacity", testCircularElemsLenGreaterThanCapacity)
	t.Run("Get", testCircularGet)
	t.Run("Peek", testCircularPeek)
	t.Run("Offer", testCircularOffer)
	t.Run("Contains", testCircularContains)
	t.Run("Clear", testCircularClear)
	t.Run("IsEmpty", testCircularIsEmpty)
	t.Run("Reset", testCircularReset)
	t.Run("Iterator", testCircularIterator)
	t.Run("MarshalJSON", testCircularMarshalJSON)
}

func testCircularCapacityOptionOverwrites(t *testing.T) {
	t.Parallel()

	circularQueue := queue.NewCircular([]int{}, 1, queue.WithCapacity(2))

	circularQueue.Offer(1)
	circularQueue.Offer(2)

	if circularQueue.Size() != 2 {
		t.Fatalf("expected size to be 2, got %d", circularQueue.Size())
	}
}

func testCircularElemsLenGreaterThanCapacity(t *testing.T) {
	t.Parallel()

	circularQueue := queue.NewCircular([]int{1, 2}, 1)

	if circularQueue.Size() != 1 {
		t.Fatalf("expected size to be 1, got %d", circularQueue.Size())
	}
}

func testCircularGet(t *testing.T) {
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
}

func testCircularPeek(t *testing.T) {
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
}

func testCircularOffer(t *testing.T) {
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
			t.Fatalf("expected elements to be %v, got %v", expectedElems, queueElems)
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
			t.Fatalf("expected elements to be %v, got %v", expectedElems, queueElems)
		}
	})
}

func testCircularContains(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		elems := []int{1, 2, 3, 4}

		circularQueue := queue.NewCircular(elems, 4)

		if !circularQueue.Contains(2) {
			t.Fatal("expected elem to be found")
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
			t.Fatal("expected elem to not be found")
		}
	})

	t.Run("EmptyQueue", func(t *testing.T) {
		circularQueue := queue.NewCircular([]int{}, 1)

		if circularQueue.Contains(1) {
			t.Fatal("expected elem to not be found")
		}
	})
}

func testCircularClear(t *testing.T) {
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
		t.Fatalf("expected elements to be %v, got %v", expectedElems, queueElems)
	}
}

func testCircularIsEmpty(t *testing.T) {
	circularQueue := queue.NewCircular([]int{}, 1)

	if !circularQueue.IsEmpty() {
		t.Fatal("expected queue to be empty")
	}
}

func testCircularReset(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

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
			t.Fatalf("expected elements to be %v, got %v", expectedElems, queueElems)
		}
	})

	t.Run("InitialElemsExceedCapacity", func(t *testing.T) {
		t.Parallel()

		// Capacity is 3 but 5 elements are provided. NewCircular trims the
		// live queue but keeps all 5 as the reset target, so Reset ends up
		// restoring size=5 on a backing array of length 3.
		circularQueue := queue.NewCircular([]int{1, 2, 3, 4, 5}, 3)

		if size := circularQueue.Size(); size != 3 {
			t.Fatalf("expected size to be 3 after construction, got %d", size)
		}

		circularQueue.Reset()

		if size := circularQueue.Size(); size != 3 {
			t.Fatalf("expected size to be 3 after Reset, got %d", size)
		}

		queueElems := circularQueue.Clear()
		expectedElems := []int{1, 2, 3}

		if !reflect.DeepEqual(expectedElems, queueElems) {
			t.Fatalf("expected elements to be %v, got %v", expectedElems, queueElems)
		}
	})
}

func testCircularIterator(t *testing.T) {
	elems := []int{1, 2, 3, 4}

	circularQueue := queue.NewCircular(elems, 5)

	iterCh := circularQueue.Iterator()

	if !circularQueue.IsEmpty() {
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

func testCircularMarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("HasElements", func(t *testing.T) {
		t.Parallel()

		elems := []int{3, 2, 1}

		q := queue.NewCircular(elems, 4)

		marshaled, err := json.Marshal(q)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedMarshaled := []byte(`[3,2,1]`)
		if !bytes.Equal(expectedMarshaled, marshaled) {
			t.Fatalf("expected marshaled to be %s, got %s", expectedMarshaled, marshaled)
		}
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()

		q := queue.NewCircular[int](nil, 1)

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
