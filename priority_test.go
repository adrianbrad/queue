package queue_test

import (
	"testing"

	"github.com/adrianbrad/queue"
	"github.com/matryer/is"
)

func TestPriority(t *testing.T) {
	t.Parallel()

	t.Run("ValidZeroValue", func(t *testing.T) {
		t.Parallel()

		i := is.New(t)

		var priorityQueue queue.Priority[intValAscending]

		_, err := priorityQueue.Get()

		i.Equal(queue.ErrNoElementsAvailable, err)

		err = priorityQueue.Offer(1)
		i.NoErr(err)

		elem, err := priorityQueue.Get()
		i.NoErr(err)
		i.Equal(intValAscending(1), elem)
	})

	t.Run("CapacityLesserThanLenElems", func(t *testing.T) {
		t.Parallel()

		i := is.New(t)

		elems := []intValAscending{4, 1, 2}

		priorityQueue := queue.NewPriority(elems, queue.WithCapacity(2))

		size := priorityQueue.Size()

		i.Equal(2, size)

		elems = drainQueue[intValAscending](priorityQueue)

		i.Equal([]intValAscending{1, 2}, elems)
	})

	t.Run("Offer", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			elems := []intValAscending{4, 1, 2}

			priorityQueue := queue.NewPriority(elems)

			err := priorityQueue.Offer(5)
			i.NoErr(err)

			size := priorityQueue.Size()

			i.Equal(4, size)

			elems = drainQueue[intValAscending](priorityQueue)

			i.Equal([]intValAscending{1, 2, 4, 5}, elems)
		})

		t.Run("ErrQueueIsFull", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			elems := []intValAscending{1}

			priorityQueue := queue.NewPriority(
				elems,
				queue.WithCapacity(1),
			)

			err := priorityQueue.Offer(2)

			i.Equal(queue.ErrQueueIsFull, err)
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			elems := []intValAscending{4, 1, 2}

			priorityQueue := queue.NewPriority(elems)

			elem, err := priorityQueue.Get()
			i.NoErr(err)

			size := priorityQueue.Size()

			i.Equal(2, size)

			i.Equal(intValAscending(1), elem)
		})

		t.Run("ErrNoElementsAvailable", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			priorityQueue := queue.NewPriority([]intValAscending{})

			_, err := priorityQueue.Get()

			i.Equal(queue.ErrNoElementsAvailable, err)
		})
	})

	t.Run("Peek", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			elems := []intValAscending{4, 1, 2}

			priorityQueue := queue.NewPriority(elems)

			elem, err := priorityQueue.Peek()
			i.NoErr(err)

			size := priorityQueue.Size()

			i.Equal(3, size)

			i.Equal(intValAscending(1), elem)

			elem, err = priorityQueue.Get()
			i.NoErr(err)

			i.Equal(intValAscending(1), elem)
		})

		t.Run("ErrNoElementsAvailable", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			priorityQueue := queue.NewPriority([]intValAscending{})

			_, err := priorityQueue.Peek()

			i.Equal(queue.ErrNoElementsAvailable, err)
		})
	})

	t.Run("Reset", func(t *testing.T) {
		t.Run("SizeGreaterThanInitialElems", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			priorityQueue := queue.NewPriority([]intValAscending{1})

			err := priorityQueue.Offer(2)
			i.NoErr(err)

			priorityQueue.Reset()

			i.Equal(1, priorityQueue.Size())
		})

		t.Run("SizeLesserThanInitialElems", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			priorityQueue := queue.NewPriority([]intValAscending{1, 2})

			_, err := priorityQueue.Get()
			i.NoErr(err)

			priorityQueue.Reset()

			i.Equal(2, priorityQueue.Size())
		})
	})
}
