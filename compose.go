package fsring

import (
	"context"
)

func ComposeErr[T, U, V any](f func(T) (U, error), g func(U) (V, error)) func(T) (V, error) {
	return func(t T) (v V, e error) {
		u, e := f(t)
		if nil != e {
			return v, e
		}
		return g(u)
	}
}

func Compose[T, U, V any](f func(T) U, g func(U) V) func(T) V {
	return func(t T) V {
		var h func(T) (V, error) = ComposeErr(
			ErrFuncGen(f),
			ErrFuncGen(g),
		)
		v, _ := h(t)
		return v
	}
}

func ComposeContext[T, U, V any](
	f func(context.Context, T) (U, error),
	g func(context.Context, U) (V, error),
) func(context.Context, T) (V, error) {
	return func(ctx context.Context, t T) (v V, e error) {
		u, e := f(ctx, t)
		if nil != e {
			return v, e
		}
		return g(ctx, u)
	}
}
