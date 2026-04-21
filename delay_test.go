package queue_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/adrianbrad/queue"
)

// Panic messages reused across test files.
const negativeCapacityPanic = "negative capacity"

type delayed struct {
	ID int       `json:"id"`
	At time.Time `json:"at"`
}

func delayedDeadline(d delayed) time.Time { return d.At }

func TestDelay(t *testing.T) {
	t.Parallel()

	t.Run("NilDeadlineFunc", testDelayNilDeadlineFunc)
	t.Run("NegativeCapacity", testDelayNegativeCapacity)
	t.Run("Get", testDelayGet)
	t.Run("GetWait", testDelayGetWait)
	t.Run("Offer", testDelayOffer)
	t.Run("OfferEarlierWakesWaiter", testDelayOfferEarlierWakesWaiter)
	t.Run("Peek", testDelayPeek)
	t.Run("Size", testDelaySize)
	t.Run("Contains", testDelayContains)
	t.Run("Clear", testDelayClear)
	t.Run("Iterator", testDelayIterator)
	t.Run("Reset", testDelayReset)
	t.Run("MarshalJSON", testDelayMarshalJSON)
	t.Run("CapacityLesserThanLenElems", testDelayCapacityLesserThanLenElems)
}

func testDelayNilDeadlineFunc(t *testing.T) {
	defer func() {
		if p := recover(); p != "nil deadline func" {
			t.Fatalf("expected panic 'nil deadline func', got %v", p)
		}
	}()

	queue.NewDelay[delayed](nil, nil)
}

func testDelayNegativeCapacity(t *testing.T) {
	t.Parallel()

	defer func() {
		if p := recover(); p != negativeCapacityPanic {
			t.Fatalf("expected panic %q, got %v", negativeCapacityPanic, p)
		}
	}()

	_ = queue.NewDelay(nil, delayedDeadline, queue.WithCapacity(-1))
}

func testDelayGet(t *testing.T) {
	t.Parallel()

	t.Run("Due", func(t *testing.T) {
		t.Parallel()

		past := time.Now().Add(-time.Minute)
		delayQueue := queue.NewDelay([]delayed{{ID: 1, At: past}}, delayedDeadline)

		got, err := delayQueue.Get()
		if err != nil {
			t.Fatalf("get: %v", err)
		}

		if got.ID != 1 {
			t.Fatalf("got id=%d want 1", got.ID)
		}
	})

	t.Run("NotDue", func(t *testing.T) {
		t.Parallel()

		future := time.Now().Add(time.Hour)
		delayQueue := queue.NewDelay([]delayed{{ID: 2, At: future}}, delayedDeadline)

		if _, err := delayQueue.Get(); !errors.Is(err, queue.ErrNoElementsAvailable) {
			t.Fatalf("expected ErrNoElementsAvailable, got %v", err)
		}
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()

		delayQueue := queue.NewDelay[delayed](nil, delayedDeadline)

		if _, err := delayQueue.Get(); !errors.Is(err, queue.ErrNoElementsAvailable) {
			t.Fatalf("expected ErrNoElementsAvailable, got %v", err)
		}
	})
}

func testDelayGetWait(t *testing.T) {
	t.Parallel()

	t.Run("WakesAfterDeadline", func(t *testing.T) {
		t.Parallel()

		due := time.Now().Add(20 * time.Millisecond)
		delayQueue := queue.NewDelay([]delayed{{ID: 1, At: due}}, delayedDeadline)

		start := time.Now()
		got := delayQueue.GetWait()
		elapsed := time.Since(start)

		if got.ID != 1 {
			t.Fatalf("got id=%d want 1", got.ID)
		}

		if elapsed < 20*time.Millisecond {
			t.Fatalf("GetWait returned before deadline: %s", elapsed)
		}
	})

	t.Run("BlocksOnEmpty", func(t *testing.T) {
		t.Parallel()

		delayQueue := queue.NewDelay[delayed](nil, delayedDeadline)

		done := make(chan delayed, 1)

		go func() {
			done <- delayQueue.GetWait()
		}()

		select {
		case <-done:
			t.Fatal("GetWait returned on empty queue")
		case <-time.After(20 * time.Millisecond):
		}

		if err := delayQueue.Offer(delayed{ID: 7, At: time.Now().Add(-time.Second)}); err != nil {
			t.Fatalf("offer: %v", err)
		}

		select {
		case got := <-done:
			if got.ID != 7 {
				t.Fatalf("got id=%d want 7", got.ID)
			}
		case <-time.After(time.Second):
			t.Fatal("GetWait did not wake after Offer")
		}
	})
}

func testDelayOffer(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		delayQueue := queue.NewDelay[delayed](nil, delayedDeadline)

		if err := delayQueue.Offer(delayed{ID: 1, At: time.Now()}); err != nil {
			t.Fatalf("offer: %v", err)
		}

		if delayQueue.Size() != 1 {
			t.Fatalf("size = %d want 1", delayQueue.Size())
		}
	})

	t.Run("ErrQueueIsFull", func(t *testing.T) {
		t.Parallel()

		delayQueue := queue.NewDelay(
			[]delayed{{ID: 1, At: time.Now()}},
			delayedDeadline,
			queue.WithCapacity(1),
		)

		err := delayQueue.Offer(delayed{ID: 2, At: time.Now()})
		if !errors.Is(err, queue.ErrQueueIsFull) {
			t.Fatalf("expected ErrQueueIsFull, got %v", err)
		}
	})
}

func testDelayOfferEarlierWakesWaiter(t *testing.T) {
	t.Parallel()

	// Head is scheduled far in the future; a new Offer with an earlier
	// deadline should cause the waiter to wake and take the new head.
	far := time.Now().Add(time.Hour)
	delayQueue := queue.NewDelay([]delayed{{ID: 1, At: far}}, delayedDeadline)

	result := make(chan delayed, 1)

	go func() {
		result <- delayQueue.GetWait()
	}()

	// Let the waiter enter its timer-wait.
	time.Sleep(20 * time.Millisecond)

	soon := time.Now().Add(10 * time.Millisecond)
	if err := delayQueue.Offer(delayed{ID: 2, At: soon}); err != nil {
		t.Fatalf("offer: %v", err)
	}

	select {
	case got := <-result:
		if got.ID != 2 {
			t.Fatalf("got id=%d want 2 (the earlier-deadline element)", got.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("waiter did not wake for earlier-deadline Offer")
	}
}

func testDelayPeek(t *testing.T) {
	t.Parallel()

	t.Run("IgnoresDeadline", func(t *testing.T) {
		t.Parallel()

		future := time.Now().Add(time.Hour)
		delayQueue := queue.NewDelay([]delayed{{ID: 7, At: future}}, delayedDeadline)

		got, err := delayQueue.Peek()
		if err != nil {
			t.Fatalf("peek: %v", err)
		}

		if got.ID != 7 {
			t.Fatalf("got id=%d want 7", got.ID)
		}

		if delayQueue.Size() != 1 {
			t.Fatal("Peek consumed an element")
		}
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()

		delayQueue := queue.NewDelay[delayed](nil, delayedDeadline)

		if _, err := delayQueue.Peek(); !errors.Is(err, queue.ErrNoElementsAvailable) {
			t.Fatalf("expected ErrNoElementsAvailable, got %v", err)
		}
	})
}

func testDelaySize(t *testing.T) {
	t.Parallel()

	delayQueue := queue.NewDelay[delayed](nil, delayedDeadline)

	if !delayQueue.IsEmpty() {
		t.Fatal("empty queue is not empty")
	}

	if err := delayQueue.Offer(delayed{ID: 1, At: time.Now()}); err != nil {
		t.Fatalf("offer: %v", err)
	}

	if delayQueue.IsEmpty() {
		t.Fatal("non-empty queue reports empty")
	}

	if delayQueue.Size() != 1 {
		t.Fatalf("size = %d want 1", delayQueue.Size())
	}
}

func testDelayContains(t *testing.T) {
	t.Parallel()

	d := delayed{ID: 42, At: time.Now().Add(time.Minute)}
	delayQueue := queue.NewDelay([]delayed{d}, delayedDeadline)

	if !delayQueue.Contains(d) {
		t.Fatal("expected queue to contain d")
	}

	if delayQueue.Contains(delayed{ID: 99, At: time.Now()}) {
		t.Fatal("expected queue to not contain a stranger")
	}
}

func testDelayClear(t *testing.T) {
	t.Parallel()

	now := time.Now()
	delayQueue := queue.NewDelay(
		[]delayed{
			{ID: 2, At: now.Add(20 * time.Millisecond)},
			{ID: 1, At: now.Add(10 * time.Millisecond)},
			{ID: 3, At: now.Add(30 * time.Millisecond)},
		},
		delayedDeadline,
	)

	cleared := delayQueue.Clear()
	ids := make([]int, len(cleared))

	for i, c := range cleared {
		ids[i] = c.ID
	}

	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(expected, ids) {
		t.Fatalf("expected %v got %v", expected, ids)
	}

	if !delayQueue.IsEmpty() {
		t.Fatal("Clear did not empty the queue")
	}
}

func testDelayIterator(t *testing.T) {
	t.Parallel()

	now := time.Now()
	delayQueue := queue.NewDelay(
		[]delayed{
			{ID: 3, At: now.Add(30 * time.Millisecond)},
			{ID: 1, At: now.Add(10 * time.Millisecond)},
			{ID: 2, At: now.Add(20 * time.Millisecond)},
		},
		delayedDeadline,
	)

	iterCh := delayQueue.Iterator()

	if !delayQueue.IsEmpty() {
		t.Fatal("Iterator did not drain the queue")
	}

	var ids []int
	for d := range iterCh {
		ids = append(ids, d.ID)
	}

	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(expected, ids) {
		t.Fatalf("expected %v got %v", expected, ids)
	}
}

func testDelayReset(t *testing.T) {
	t.Parallel()

	base := time.Now()
	initial := []delayed{
		{ID: 1, At: base.Add(time.Minute)},
		{ID: 2, At: base.Add(2 * time.Minute)},
	}

	delayQueue := queue.NewDelay(initial, delayedDeadline)

	if err := delayQueue.Offer(delayed{ID: 99, At: base.Add(time.Second)}); err != nil {
		t.Fatalf("offer: %v", err)
	}

	delayQueue.Reset()

	if delayQueue.Size() != len(initial) {
		t.Fatalf("size after Reset = %d want %d", delayQueue.Size(), len(initial))
	}

	if delayQueue.Contains(delayed{ID: 99, At: base.Add(time.Second)}) {
		t.Fatal("Reset did not drop Offered element")
	}
}

func testDelayMarshalJSON(t *testing.T) {
	t.Parallel()

	now := time.Now()
	delayQueue := queue.NewDelay(
		[]delayed{
			{ID: 3, At: now.Add(30 * time.Millisecond)},
			{ID: 1, At: now.Add(10 * time.Millisecond)},
			{ID: 2, At: now.Add(20 * time.Millisecond)},
		},
		delayedDeadline,
	)

	marshaled, err := json.Marshal(delayQueue)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got []struct {
		ID int       `json:"id"`
		At time.Time `json:"at"`
	}

	if err := json.Unmarshal(marshaled, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	ids := []int{got[0].ID, got[1].ID, got[2].ID}
	expected := []int{1, 2, 3}

	if !reflect.DeepEqual(expected, ids) {
		t.Fatalf("expected ids %v got %v (raw: %s)", expected, ids, marshaled)
	}

	// Re-marshal to confirm it's stable.
	second, err := json.Marshal(delayQueue)
	if err != nil {
		t.Fatalf("marshal 2: %v", err)
	}

	if !bytes.Equal(marshaled, second) {
		t.Fatalf("second marshal differs: %s vs %s", marshaled, second)
	}
}

func testDelayCapacityLesserThanLenElems(t *testing.T) {
	t.Parallel()

	delayQueue := queue.NewDelay(
		[]delayed{
			{ID: 1, At: time.Now()},
			{ID: 2, At: time.Now()},
			{ID: 3, At: time.Now()},
		},
		delayedDeadline,
		queue.WithCapacity(2),
	)

	if delayQueue.Size() != 2 {
		t.Fatalf("expected size 2 after construction, got %d", delayQueue.Size())
	}
}

func BenchmarkDelayQueue(b *testing.B) {
	past := time.Now().Add(-time.Hour)

	b.Run("Peek", func(b *testing.B) {
		delayQueue := queue.NewDelay([]delayed{{ID: 1, At: past}}, delayedDeadline)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_, _ = delayQueue.Peek()
		}
	})

	b.Run("Get_Offer", func(b *testing.B) {
		delayQueue := queue.NewDelay([]delayed{{ID: 1, At: past}}, delayedDeadline)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_, _ = delayQueue.Get()

			_ = delayQueue.Offer(delayed{ID: i, At: past})
		}
	})

	b.Run("Offer", func(b *testing.B) {
		delayQueue := queue.NewDelay[delayed](nil, delayedDeadline)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_ = delayQueue.Offer(delayed{ID: i, At: past})
		}
	})
}

// Race: concurrent Offers + GetWaits eventually drain all elements.
func TestDelayConcurrentOfferAndGet(t *testing.T) {
	t.Parallel()

	const (
		producers    = 8
		perProducer  = 25
		totalDelayed = producers * perProducer
	)

	delayQueue := queue.NewDelay[delayed](nil, delayedDeadline)

	var offered sync.WaitGroup

	offered.Add(producers)

	for p := 0; p < producers; p++ {
		go func(p int) {
			defer offered.Done()

			for i := 0; i < perProducer; i++ {
				when := time.Now().Add(time.Duration(i%5) * time.Millisecond)

				if err := delayQueue.Offer(delayed{ID: p*perProducer + i, At: when}); err != nil {
					t.Errorf("offer: %v", err)
					return
				}
			}
		}(p)
	}

	seen := make(chan int, totalDelayed)

	var consumers sync.WaitGroup

	consumers.Add(producers)

	for c := 0; c < producers; c++ {
		go func() {
			defer consumers.Done()

			for i := 0; i < perProducer; i++ {
				got := delayQueue.GetWait()
				seen <- got.ID
			}
		}()
	}

	offered.Wait()
	consumers.Wait()

	close(seen)

	ids := make([]int, 0, totalDelayed)
	for id := range seen {
		ids = append(ids, id)
	}

	sort.Ints(ids)

	for i := 0; i < totalDelayed; i++ {
		if ids[i] != i {
			t.Fatalf("missing or duplicate id at position %d: %d", i, ids[i])
		}
	}
}
