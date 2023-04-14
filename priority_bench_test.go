package queue_test

import (
	"testing"

	"github.com/adrianbrad/queue"
)

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
