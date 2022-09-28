package fsring

const OptHasValue = true
const OptEmpty = false

func OptMap[T, U any](t T, hasValue bool, f func(T) U) (u U, b bool) {
	if hasValue {
		return f(t), OptHasValue
	}
	return u, OptEmpty
}
