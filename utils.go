package queue

// noopLesser is used by structs that depend on the Lesser interface
// as one of their type parameters, in order to enable asserting
// that those structs implement different interfaces.
type noopLesser struct{}

// Less always returns false.
func (noopLesser) Less(any) bool { return false }
