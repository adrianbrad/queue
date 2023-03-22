package queue_test

import (
	"testing"

	"github.com/adrianbrad/queue"
)

func BenchmarkBlockingQueue(b *testing.B) {
	b.Run("Peek", func(b *testing.B) {
		blockingQueue := queue.NewBlocking([]int{1})

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_, _ = blockingQueue.Peek()
		}
	})

	b.Run("Get_Offer", func(b *testing.B) {
		blockingQueue := queue.NewBlocking([]int{1})

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_, _ = blockingQueue.Get()

			_ = blockingQueue.Offer(1)
		}
	})

	b.Run("Offer", func(b *testing.B) {
		blockingQueue := queue.NewBlocking[int](nil)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i <= b.N; i++ {
			_ = blockingQueue.Offer(i)
		}
	})
}
