package fsring

import (
	"errors"
	"io/fs"
	"os"
)

type Read func(filename string) (data []byte, e error)

type ReadByUint[T uint8 | uint16] func(key T) (data []byte, e error)

func (r ReadByUint[T]) orElse(ef func(error) ([]byte, error)) ReadByUint[T] {
	return ErrOrElseGen(r, ef)
}

func (r ReadByUint[T]) ErrorIgnored(ignoreMe error) ReadByUint[T] {
	ef := func(e error) ([]byte, error) {
		return ErrFromBool(
			errors.Is(e, ignoreMe),
			func() []byte { return nil },
			func() error { return e },
		)
	}
	return r.orElse(ef)
}

func (r ReadByUint[T]) NoentIgnored() ReadByUint[T] { return r.ErrorIgnored(fs.ErrNotExist) }

func ReadByUintBuilder[T uint8 | uint16](bld NameBuilderUint[T]) func(NameChecker) ReadByUint[T] {
	return func(chk NameChecker) ReadByUint[T] {
		var fullpath2data func(unchecked string) ([]byte, error) = ComposeErr(
			func(unchecked string) (checked string, e error) { return chk(unchecked), nil },
			os.ReadFile,
		)
		return ComposeErr(
			ErrFuncGen(bld), // T -> string, nil
			fullpath2data,   // string -> []byte, error
		)
	}
}
