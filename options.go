package queue

type options struct {
	capacity *int
}

// An Option configures a Queue using the functional options paradigm.
type Option interface {
	apply(*options)
}

type capacityOption int

func (c capacityOption) apply(opts *options) {
	ic := int(c)

	opts.capacity = &ic
}

// WithCapacity specifies a fixed capacity for a queue.
func WithCapacity(capacity int) Option {
	return capacityOption(capacity)
}
