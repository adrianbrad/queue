package queue

import "testing"

func TestBlockingGetZerosPoppedSlot(t *testing.T) {
	type payload struct{ id int }

	q := NewBlocking[*payload](nil)
	if err := q.Offer(&payload{id: 1}); err != nil {
		t.Fatalf("offer: %v", err)
	}

	backing := q.elems[:cap(q.elems)]
	got, err := q.Get()
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil || got.id != 1 {
		t.Fatalf("unexpected element: %v", got)
	}
	if len(q.elems) != 0 {
		t.Fatalf("expected empty queue after Get, got size %d", len(q.elems))
	}
	if backing[0] != nil {
		t.Fatal("backing slot still retains the popped payload")
	}
}

func TestBlockingClearZerosBackingSlots(t *testing.T) {
	type payload struct{ id int }

	q := NewBlocking[*payload](nil)
	for i := 0; i < 3; i++ {
		if err := q.Offer(&payload{id: i + 1}); err != nil {
			t.Fatalf("offer: %v", err)
		}
	}

	backing := q.elems[:cap(q.elems)]
	removed := q.Clear()
	if len(removed) != 3 {
		t.Fatalf("expected 3 removed elements, got %d", len(removed))
	}
	if q.Size() != 0 {
		t.Fatalf("expected empty queue after clear, got size %d", q.Size())
	}

	for i, elem := range backing {
		if elem != nil {
			t.Fatalf("backing slot %d still retains a payload after Clear", i)
		}
	}
}
