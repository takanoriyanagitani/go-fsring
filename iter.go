package fsring

type Iter[T any] func() (t T, hasValue bool)

func IterReduce[T, U any](i Iter[T], init U, reducer func(state U, t T) U) U {
	var state U = init
	for o, hasValue := i(); hasValue; o, hasValue = i() {
		var t T = o
		state = reducer(state, t)
	}
	return state
}

func IterFromArr[T any](a []T) Iter[T] {
	var ix int = 0
	return func() (t T, hasValue bool) {
		if ix < len(a) {
			t = a[ix]
			ix += 1
			return t, OptHasValue
		}
		return t, OptEmpty
	}
}

func (i Iter[T]) TryForEach(f func(T) error) error {
	return IterReduce(i, nil, func(state error, t T) error {
		return ErrOrElse(state, func() error { return f(t) })
	})
}

func (i Iter[T]) All(f func(T) bool) bool {
	return IterReduce(i, true, func(state bool, t T) bool {
		return state && f(t)
	})
}

func IterEmpty[T any]() Iter[T] {
	return func() (t T, hasValue bool) {
		return t, OptEmpty
	}
}

func (i Iter[T]) Count() int {
	return IterReduce(i, 0, func(state int, _ T) int {
		return state + 1
	})
}

func (i Iter[T]) ToArray() []T {
	return IterReduce(i, nil, func(state []T, t T) []T { return append(state, t) })
}

func (i Iter[T]) ToArrayIter() Iter[T] { return IterFromArr(i.ToArray()) }

func IterMap[T, U any](i Iter[T], f func(T) U) Iter[U] {
	return func() (u U, hasValue bool) {
		t, hasValue := i()
		return OptMap(t, hasValue, f)
	}
}

// Filter creates filtered Iter from filtered array.
func (i Iter[T]) Filter(f func(T) (ok bool)) Iter[T] {
	var a2a func([]T) []T = Curry(ArrayFilter[T])(f)
	return Compose(
		func(_ Iter[T]) []T { return i.ToArray() },
		Compose(
			a2a,
			IterFromArr[T],
		),
	)(i)
}

func (i Iter[T]) Reduce(init T, reducer func(state T, item T) T) T {
	return IterReduce(i, init, reducer)
}

func IterInts(lbi, ube int) Iter[int] {
	var ix int = lbi
	return func() (i int, hasValue bool) {
		if ix < ube {
			var nx int = ix
			ix += 1
			return nx, OptHasValue
		}
		return -1, OptEmpty
	}
}
