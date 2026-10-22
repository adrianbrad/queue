package queue

import "testing"

func TestPriorityGetZerosPoppedSlot(t *testing.T) {
	type payload struct{ id int }

	q := NewPriority([]*payload{{id: 1}}, func(a, b *payload) bool {
		return a.id < b.id
	})
	backing := q.elements.elems[:cap(q.elements.elems)]

	got, err := q.Get()
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil || got.id != 1 {
		t.Fatalf("unexpected element: %v", got)
	}
	if len(q.elements.elems) != 0 {
		t.Fatalf("expected empty queue after Get, got size %d", len(q.elements.elems))
	}
	if backing[0] != nil {
		t.Fatal("backing slot still retains the popped payload")
	}
}
