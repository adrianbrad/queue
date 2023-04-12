package queue_test

import (
	"testing"

	"github.com/adrianbrad/queue"
)

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
