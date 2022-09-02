package queue_test

import (
	"bytes"
	"github.com/adrianbrad/queue"
	"testing"
)

// goos: darwin
// goarch: amd64
// pkg: github.com/adrianbrad/queue
// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
// BenchmarkBlocking_1-12          11742712               103.1 ns/op           120 B/op          3 allocs/op
// BenchmarkBlocking2_1-12         11117152               109.2 ns/op           120 B/op          2 allocs/op
// BenchmarkBlocking_10-12         11871397               103.1 ns/op           120 B/op          3 allocs/op
// BenchmarkBlocking2_10-12         3911952               310.5 ns/op           120 B/op          2 allocs/op

func BenchmarkBlocking_1(b *testing.B) {
	elems := bytes.Repeat([]byte("a"), 1)
	q := queue.NewBlocking(elems)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q.Reset()
	}
}

func BenchmarkBlocking2_1(b *testing.B) {
	elems := bytes.Repeat([]byte("a"), 1)
	q := queue.NewBlocking2(elems)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q.Reset()
	}
}

func BenchmarkBlocking_10(b *testing.B) {
	elems := bytes.Repeat([]byte("a"), 10)
	q := queue.NewBlocking(elems)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q.Reset()
	}
}

func BenchmarkBlocking2_10(b *testing.B) {
	elems := bytes.Repeat([]byte("a"), 10)
	q := queue.NewBlocking2(elems)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q.Reset()
	}
}
