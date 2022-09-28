package fsring

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

var ErrTooManyQueue = errors.New("too many queue already")

type Next func(dirname string) (next string, e error)

func NextBuilderConst(filename string) Next {
	return func(dirname string) (next string, e error) {
		return filepath.Join(dirname, filename), nil
	}
}

func NextBuilderHeavyBuilderNewU[T uint8 | uint16](max int, format string) func(IsEmpty) Next {
	return func(ie IsEmpty) Next {
		return func(dirname string) (next string, e error) {
			for i := 0; i < max; i++ {
				next = filepath.Join(dirname, fmt.Sprintf(format, i))
				empty, e := ie(next)
				if nil != e {
					return "", e
				}
				if empty {
					return next, nil
				}
			}
			return "", ErrTooManyQueue
		}
	}
}

var NextBuilderHeavy3 func(IsEmpty) Next = NextBuilderHeavyBuilderNewU[uint8](256, "%02x")
var NextBuilderHeavy4 func(IsEmpty) Next = NextBuilderHeavyBuilderNewU[uint8](65536, "%04x")

type NextName func(prev string) (next string, e error)

func (n NextName) BasenameOnly() NextName {
	return func(prev string) (next string, e error) {
		var bp string = filepath.Base(prev)
		var dp string = filepath.Dir(prev)
		return ComposeErr(
			n, // basename string -> base next string, error
			func(baseNext string) (string, error) { return filepath.Join(dp, baseNext), nil },
		)(bp)
	}
}

func (n Next) Fallback(other Next, check func(error) error) Next {
	return func(dirname string) (next string, e error) {
		next, e = n(dirname)
		if nil != e {
			e = check(e)
			if nil != e {
				return "", e
			}
			return other(dirname)
		}
		return next, e
	}
}

func (n Next) FallbackIfNotEmpty(ie IsEmpty, other Next) Next {
	var checkEmpty func(dirname string) (empty bool, e error) = ComposeErr(
		n,
		ie,
	)
	return func(dirname string) (next string, e error) {
		empty, e := checkEmpty(dirname)
		if nil != e {
			return "", e
		}
		return SelectFunc(other, n)(empty)(dirname)
	}
}

func NextBuilderNew(chk NameChecker) func(dirname string) func(managename string) Next {
	return func(dirname string) func(managename string) Next {
		return func(managefilename string) Next {
			return func(dirname string) (next string, e error) {
				var fullpath string = filepath.Join(dirname, managefilename)
				return ComposeErr(
					os.ReadFile, // string -> []byte, error
					func(content []byte) (string, error) { return string(content), nil },
				)(chk(fullpath))
			}
		}
	}
}

type NextNameU[T uint8 | uint16] func(prev T) (next T, e error)

func NextNameUdefault[T uint8 | uint16](prev T) (next T, e error) { return prev + 1, nil }

func (n NextNameU[T]) ToNextName(sz int, format string) NextName {
	var nex NextName = ComposeErr(
		uintParserHexBuilder[T](sz),
		ComposeErr(
			n, // T -> T, err
			func(next T) (string, error) {
				return fmt.Sprintf(format, next), nil
			},
		),
	)
	return nex.BasenameOnly()
}

var Next3default NextNameU[uint8] = NextNameUdefault[uint8]
var Next4default NextNameU[uint16] = NextNameUdefault[uint16]

func newUintParser(base int, bitSize int) func(s string) (uint64, error) {
	return func(s string) (uint64, error) {
		return strconv.ParseUint(s, base, bitSize)
	}
}

func uintParserHexBuilder[T uint8 | uint16](bitsz int) func(s string) (T, error) {
	return ComposeErr(
		newUintParser(16, bitsz),
		func(u uint64) (T, error) { return T(u), nil },
	)
}

var NextNameDefault4 NextName = Next4default.ToNextName(16, "%04x")
var NextNameDefault3 NextName = Next3default.ToNextName(8, "%02x")
