package queue

import (
	"testing"

	"github.com/matryer/is"
)

func TestNoopLesser(t *testing.T) {
	t.Parallel()

	i := is.New(t)

	var l noopLesser

	i.True(!l.Less(1))
	i.True(!l.Less("a"))
}
