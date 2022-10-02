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

func OptFromBool[T any](b bool, f func() T) (t T, hasValue bool) {
	if b {
		return OptNew(f())
	}
	return t, OptEmpty
}

func OptUnwrapOrElse[T any](f func() (T, bool), alt func() T) func() T {
	return func() T {
		t, hasValue := f()
		if hasValue {
			return t
		}
		return alt()
	}
}

func OptUnwrapOrDefault[T any](f func() (T, bool)) func() T {
	return OptUnwrapOrElse(
		f,
		func() (t T) { return },
	)
}
