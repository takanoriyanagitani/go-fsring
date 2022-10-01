package fsring

import (
	"os"
)

type Read func(filename string) (data []byte, e error)

type ReadByUint[T uint8 | uint16] func(key T) (data []byte, e error)

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
