package fsring

import (
	"errors"
)

func ErrOrElse(e error, ef func() error) error {
	if nil != e {
		return e
	}
	return ef()
}

func Err1st(ef []func() error) error {
	var ei Iter[func() error] = IterFromArr(ef)
	return IterReduce(ei, nil, ErrOrElse)
}

func ErrFuncGen[T, U any](f func(T) U) func(T) (U, error) {
	return func(t T) (U, error) {
		return f(t), nil
	}
}

func ErrOnly[T, U any](f func(T) (U, error)) func(T) error {
	return func(t T) error {
		_, e := f(t)
		return e
	}
}

func ErrFromBool[T any](ok bool, okf func() T, ngf func() error) (t T, e error) {
	if ok {
		return okf(), nil
	}
	return t, ngf()
}

func ErrUnwrapOrElse[T, U any](f func(T) (U, error), g func(error) U) func(T) U {
	return func(t T) U {
		u, e := f(t)
		if nil != e {
			return g(e)
		}
		return u
	}
}

func ErrIgnore(check func(error) (ignore bool)) func(error) error {
	return func(e error) error {
		if check(e) {
			return nil
		}
		return e
	}
}

func ErrIgnoreNew(e error) func(error) error {
	return ErrIgnore(func(other error) (ignore bool) {
		return errors.Is(other, e)
	})
}

func ErrIgnored[T any](f func(T) error, check func(error) (ignore bool)) func(T) error {
	return func(t T) error {
		e := f(t)
		if nil != e {
			var ignore bool = check(e)
			if ignore {
				return nil
			}
			return e
		}
		return nil
	}
}

func ErrOrElseGen[T, U any](f func(T) (U, error), ef func(error) (U, error)) func(T) (U, error) {
	return func(t T) (U, error) {
		u, e := f(t)
		if nil != e {
			return ef(e)
		}
		return u, nil
	}
}

func ErrFallback[T, U any](f func(T) (U, error), g func(T) (U, error)) func(T) (U, error) {
	return func(t T) (U, error) {
		u, e := f(t)
		if nil != e {
			return g(t)
		}
		return u, e
	}
}
