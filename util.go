package fsring

func CurryErr[T, U, V any](f func(T, U) (V, error)) func(T) func(U) (V, error) {
	return func(t T) func(U) (V, error) {
		return func(u U) (V, error) {
			return f(t, u)
		}
	}
}

func Curry[T, U, V any](f func(T, U) V) func(T) func(U) V {
	return func(t T) func(U) V {
		return func(u U) V {
			return f(t, u)
		}
	}
}

func IgnoreArg[T, U any](f func() (U, error)) func(T) (U, error) {
	return func(_ T) (U, error) {
		return f()
	}
}

func Identity[T any](t T) T { return t }

func SelectFunc[T, U any](f func(T) (U, error), g func(T) (U, error)) func(latter bool) func(T) (U, error) {
	return func(latter bool) func(T) (U, error) {
		if latter {
			return g
		}
		return f
	}
}

func Partial[T, U any](f func(T) U, t T) func() U { return func() U { return f(t) } }

func CurryErrIII[T, U, V, W any](f func(T, U, V) (W, error)) func(T) func(U) func(V) (W, error) {
	return func(t T) func(U) func(V) (W, error) {
		g := func(u U, v V) (W, error) {
			return f(t, u, v)
		}
		return CurryErr(g)
	}
}
