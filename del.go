package fsring

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type DeleteUint[T uint8 | uint16] func(target T) error

func (d DeleteUint[T]) errIgnored(check func(error) (ignore bool)) DeleteUint[T] {
	return ErrIgnored(d, check)
}

func (d DeleteUint[T]) ErrIgnored(ignoreMe error) DeleteUint[T] {
	return d.errIgnored(func(e error) (ignore bool) {
		return errors.Is(e, ignoreMe)
	})
}

func (d DeleteUint[T]) NoentIgnored() DeleteUint[T] { return d.ErrIgnored(fs.ErrNotExist) }

func truncateBuilder(chk NameChecker) func(fullpath string) error {
	return func(fullpath string) error {
		return os.Truncate(chk(fullpath), 0)
	}
}

type NameBuilderUint[T uint8 | uint16] func(t T) (fullpath string)

func NameBuilderUintNew[T uint8 | uint16](u2h uint2hex[T]) func(dirname string) NameBuilderUint[T] {
	return func(dirname string) NameBuilderUint[T] {
		return Compose(
			u2h, // T -> string
			func(h string) string { return filepath.Join(dirname, h) },
		)
	}
}

var NameBuilderUint3 func(dirname string) NameBuilderUint[uint8] = NameBuilderUintNew(uint2hex3)
var NameBuilderUint4 func(dirname string) NameBuilderUint[uint16] = NameBuilderUintNew(uint2hex4)

func DeleteUintBuilder[T uint8 | uint16](bld NameBuilderUint[T]) func(NameChecker) DeleteUint[T] {
	return func(chk NameChecker) DeleteUint[T] {
		var path2truncate func(fullpath string) error = truncateBuilder(chk)
		var f func(t T) (string, error) = ComposeErr(
			ErrFuncGen(bld), // T -> string, error
			func(fullpath string) (string, error) { return "", path2truncate(fullpath) },
		)
		return ErrOnly(f)
	}
}
