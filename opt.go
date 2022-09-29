package fsring

const OptHasValue = true
const OptEmpty = false

func OptNew[T any](t T) (T, bool) { return t, OptHasValue }

func OptMap[T, U any](t T, hasValue bool, f func(T) U) (u U, b bool) {
	if hasValue {
		return OptNew(f(t))
	}
	return u, OptEmpty
}
