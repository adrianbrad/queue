package queue_test

import (
	"fmt"
	"testing"

	"github.com/adrianbrad/queue"
	"github.com/matryer/is"
)

func TestPriority(t *testing.T) {
	t.Parallel()

	lessAscending := func(elem, elemAfter int) bool {
		return elem < elemAfter
	}

	lessInt := func(elem, elemAfter int) bool {
		return elem < elemAfter
	}

	t.Run("ValidZeroValue", func(t *testing.T) {
		t.Parallel()

		i := is.New(t)

		var priorityQueue queue.Priority[int]

		_, err := priorityQueue.Get()

		i.Equal(queue.ErrNoElementsAvailable, err)

		err = priorityQueue.Offer(1)
		i.NoErr(err)

		elem, err := priorityQueue.Get()
		i.NoErr(err)
		i.Equal(int(1), elem)
	})

	t.Run("CapacityLesserThanLenElems", func(t *testing.T) {
		t.Parallel()

		i := is.New(t)

		elems := []int{4, 1, 2}

		priorityQueue := queue.NewPriority(elems, lessInt, queue.WithCapacity(2))

		size := priorityQueue.Size()

		i.Equal(2, size)

		elems = drainQueue[int](priorityQueue)

		i.Equal([]int{1, 2}, elems)
	})

	t.Run("Offer", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			elems := []int{4, 1, 2}

			priorityQueue := queue.NewPriority(elems, lessAscending)

			err := priorityQueue.Offer(5)
			i.NoErr(err)

			size := priorityQueue.Size()

			i.Equal(4, size)

			drainedElems := drainQueue[int](priorityQueue)

			i.Equal([]int{1, 2, 4, 5}, drainedElems)

			newElems := make([]int, 10)

			for j := 19; j >= 10; j-- {
				newElems[j%10] = j

				err := priorityQueue.Offer(j)
				i.NoErr(err)
			}

			drainedElems = drainQueue[int](priorityQueue)
			i.Equal(newElems, drainedElems)
		})

		t.Run("ErrQueueIsFull", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			elems := []int{1}

			priorityQueue := queue.NewPriority(
				elems, lessAscending,
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

			elems := []int{4, 1, 2}

			priorityQueue := queue.NewPriority(elems, lessInt)

			elem, err := priorityQueue.Get()
			i.NoErr(err)

			size := priorityQueue.Size()

			i.Equal(2, size)

			i.Equal(1, elem)
		})

		t.Run("ErrNoElementsAvailable", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			priorityQueue := queue.NewPriority([]int{}, nil)

			_, err := priorityQueue.Get()

			i.Equal(queue.ErrNoElementsAvailable, err)
		})
	})

	t.Run("Peek", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			elems := []int{4, 1, 2}

			priorityQueue := queue.NewPriority(elems, lessAscending)

			elem, err := priorityQueue.Peek()
			i.NoErr(err)

			size := priorityQueue.Size()

			i.Equal(3, size)

			i.Equal(1, elem)

			elem, err = priorityQueue.Get()
			i.NoErr(err)

			i.Equal(1, elem)

			elem, err = priorityQueue.Peek()
			i.NoErr(err)

			size = priorityQueue.Size()

			i.Equal(2, size)

			i.Equal(2, elem)

			elem, err = priorityQueue.Get()
			i.NoErr(err)

			i.Equal(2, elem)
		})

		t.Run("ErrNoElementsAvailable", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			priorityQueue := queue.NewPriority([]int{}, nil)

			_, err := priorityQueue.Peek()

			i.Equal(queue.ErrNoElementsAvailable, err)
		})
	})

	t.Run("Reset", func(t *testing.T) {
		t.Run("SizeGreaterThanInitialElems", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			elems := []int{1}

			priorityQueue := queue.NewPriority(elems, lessAscending)

			err := priorityQueue.Offer(2)
			i.NoErr(err)

			priorityQueue.Reset()

			i.Equal(1, priorityQueue.Size())
		})

		t.Run("SizeLesserThanInitialElems", func(t *testing.T) {
			t.Parallel()

			i := is.New(t)

			elems := []int{1, 2}

			priorityQueue := queue.NewPriority(elems, lessAscending)

			_, err := priorityQueue.Get()
			i.NoErr(err)

			priorityQueue.Reset()

			i.Equal(2, priorityQueue.Size())
		})
	})
}

func FuzzPriority(f *testing.F) {
	testcases := [][]byte{{1, 2, 3}, {4, 5, 6}, {9, 8, 7}}
	for _, tc := range testcases {
		f.Add(tc)
	}

	lessFunc := func(elem, elemAfter byte) bool {
		return elem < elemAfter
	}

	testFunc := func(t *testing.T, priorityQueue *queue.Priority[byte], orig []byte) {
		i := is.New(t)

		for _, v := range orig {
			err := priorityQueue.Offer(v)
			i.NoErr(err)
		}

		for _, v := range orig {
			t.Logf("v: %d", v)

			peekedVal, err := priorityQueue.Peek()
			i.NoErr(err)

			i.Equal(v, peekedVal)

			getVal, err := priorityQueue.Get()
			i.NoErr(err)

			i.Equal(peekedVal, getVal)
		}
	}

	f.Fuzz(func(t *testing.T, orig []byte) {
		t.Run("ZeroValue", func(t *testing.T) {
			// t.Parallel()

			priorityQueue := &queue.Priority[byte]{}

			testFunc(t, priorityQueue, orig)
		})

		t.Run("Constructor", func(t *testing.T) {
			// t.Parallel()

			priorityQueue := queue.NewPriority(nil, lessFunc)

			testFunc(t, priorityQueue, orig)
		})
	})
}

func TestZeroVal(t *testing.T) {
	vs := []byte{1, 2, 3, 4}

	q := &queue.Priority[byte]{}

	i := is.New(t)

	for _, v := range vs {
		err := q.Offer(v)
		i.NoErr(err)
	}

	fmt.Println(q.Peek())
	fmt.Println(q.Get())
	fmt.Println(q.Peek())
	fmt.Println(q.Get())
	fmt.Println(q.Peek())
	fmt.Println(q.Get())
	fmt.Println(q.Peek())
	fmt.Println(q.Get())

	q = queue.NewPriority[byte](vs, func(elem, elemAfter byte) bool {
		return elem < elemAfter
	})

	fmt.Println(q.Peek())
	fmt.Println(q.Get())
	fmt.Println(q.Peek())
	fmt.Println(q.Get())
}
