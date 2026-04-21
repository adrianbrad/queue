package queue_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/adrianbrad/queue"
)

func TestLinked(t *testing.T) {
	t.Parallel()

	t.Run("Get", testLinkedGet)
	t.Run("Peek", testLinkedPeek)
	t.Run("Offer", testLinkedOffer)
	t.Run("Contains", testLinkedContains)
	t.Run("Clear", testLinkedClear)
	t.Run("IsEmpty", testLinkedIsEmpty)
	t.Run("Reset", testLinkedReset)
	t.Run("Iterator", testLinkedIterator)
	t.Run("MarshalJSON", testLinkedMarshalJSON)
	t.Run("OfferReusesPoppedNode", testLinkedOfferReusesPoppedNode)
	t.Run("DrainBeyondFreeCap", testLinkedDrainBeyondFreeCap)
}

// testLinkedOfferReusesPoppedNode drives the free-list-has-node branch
// of the internal node recycling.
func testLinkedOfferReusesPoppedNode(t *testing.T) {
	t.Parallel()

	linkedQueue := queue.NewLinked([]int{1, 2, 3})

	got, err := linkedQueue.Get()
	if err != nil || got != 1 {
		t.Fatalf("get: %d %v", got, err)
	}

	// This Offer reuses the node just released by Get.
	if err := linkedQueue.Offer(4); err != nil {
		t.Fatalf("offer: %v", err)
	}

	cleared := linkedQueue.Clear()
	expected := []int{2, 3, 4}

	if !reflect.DeepEqual(expected, cleared) {
		t.Fatalf("expected %v got %v", expected, cleared)
	}
}

// testLinkedDrainBeyondFreeCap drives the cap-reached branch of the
// internal recycle() so the free list stops growing past its cap.
func testLinkedDrainBeyondFreeCap(t *testing.T) {
	t.Parallel()

	// Use a size that exceeds the free-list cap so the cap-reached
	// branch of recycle() executes.
	const n = 128

	linkedQueue := queue.NewLinked[int](nil)

	for i := 0; i < n; i++ {
		if err := linkedQueue.Offer(i); err != nil {
			t.Fatalf("offer %d: %v", i, err)
		}
	}

	for i := 0; i < n; i++ {
		got, err := linkedQueue.Get()
		if err != nil {
			t.Fatalf("get %d: %v", i, err)
		}

		if got != i {
			t.Fatalf("get %d: want %d got %d", i, i, got)
		}
	}

	if !linkedQueue.IsEmpty() {
		t.Fatal("queue should be empty")
	}
}

func testLinkedGet(t *testing.T) {
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
}

func testLinkedPeek(t *testing.T) {
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
}

func testLinkedOffer(t *testing.T) {
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
}

func testLinkedContains(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		elems := []int{1, 2, 3, 4}

		linkedQueue := queue.NewLinked(elems)

		if !linkedQueue.Contains(2) {
			t.Fatal("expected elem to be found")
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
			t.Fatal("expected elem to not be found")
		}
	})

	t.Run("EmptyQueue", func(t *testing.T) {
		linkedQueue := queue.NewLinked([]int{})

		if linkedQueue.Contains(1) {
			t.Fatal("expected elem to not be found")
		}
	})
}

func testLinkedClear(t *testing.T) {
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
}

func testLinkedIsEmpty(t *testing.T) {
	linkedQueue := queue.NewLinked([]int{})

	if !linkedQueue.IsEmpty() {
		t.Fatal("expected queue to be empty")
	}
}

func testLinkedReset(t *testing.T) {
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
}

func testLinkedIterator(t *testing.T) {
	elems := []int{1, 2, 3, 4}

	linkedQueue := queue.NewLinked(elems)

	iterCh := linkedQueue.Iterator()

	if !linkedQueue.IsEmpty() {
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

func testLinkedMarshalJSON(t *testing.T) {
	t.Parallel()

	elems := []int{3, 2, 1}

	q := queue.NewLinked(elems)

	marshaled, err := json.Marshal(q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedMarshaled := []byte(`[3,2,1]`)
	if !bytes.Equal(expectedMarshaled, marshaled) {
		t.Fatalf("expected marshaled to be %s, got %s", expectedMarshaled, marshaled)
	}
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
