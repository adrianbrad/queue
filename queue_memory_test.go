package queue

import "testing"

func TestCircularResetZerosSlotsBeyondInitialElements(t *testing.T) {
	type payload struct{ id int }

	q := NewCircular([]*payload{{id: 1}, {id: 2}}, 5)
	for i := 3; i <= 5; i++ {
		if err := q.Offer(&payload{id: i}); err != nil {
			t.Fatalf("offer: %v", err)
		}
	}

	q.Reset()
	if q.size != 2 {
		t.Fatalf("expected reset size 2, got %d", q.size)
	}
	for i := 2; i < len(q.elems); i++ {
		if q.elems[i] != nil {
			t.Fatalf("slot %d still retains an extra payload after Reset", i)
		}
	}
}

func TestPriorityResetDropsOversizedBacking(t *testing.T) {
	type payload struct{ id int }

	lessByID := func(a, b *payload) bool { return a.id < b.id }
	q := NewPriority([]*payload{{id: 1}, {id: 2}}, lessByID)
	for i := 3; i <= 7; i++ {
		if err := q.Offer(&payload{id: i}); err != nil {
			t.Fatalf("offer: %v", err)
		}
	}
	if cap(q.elements.elems) <= 2 {
		t.Fatalf("expected offered items to grow the backing array, got capacity %d", cap(q.elements.elems))
	}

	q.Reset()
	if len(q.elements.elems) != 2 || cap(q.elements.elems) != 2 {
		t.Fatalf("expected reset to retain only the two initial slots, got len=%d cap=%d", len(q.elements.elems), cap(q.elements.elems))
	}
	for _, elem := range q.elements.elems {
		if elem == nil || elem.id < 1 || elem.id > 2 {
			t.Fatalf("unexpected element after Reset: %v", elem)
		}
	}
}
