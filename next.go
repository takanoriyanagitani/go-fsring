package fsring

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type Next func(dirname string) (next string, e error)

func NextBuilderConst(filename string) Next {
	return func(dirname string) (next string, e error) {
		return filepath.Join(dirname, filename), nil
	}
}

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

type Next4 func(prev uint16) (next uint16, e error)

func (n Next4) ToNextName() NextName {
	var nex NextName = ComposeErr(
		Uint16parserHex, // basename string -> uint16, error
		ComposeErr(
			n, // uint16 -> uint16, err
			func(next uint16) (string, error) {
				return fmt.Sprintf("%04x", next), nil
			},
		),
	)
	return nex.BasenameOnly()
}

var Next4default Next4 = func(prev uint16) (next uint16, e error) { return prev + 1, nil }

func newUintParser(base int, bitSize int) func(s string) (uint64, error) {
	return func(s string) (uint64, error) {
		return strconv.ParseUint(s, base, bitSize)
	}
}

var Uint16parserHex func(s string) (uint16, error) = ComposeErr(
	newUintParser(16, 16),
	func(u uint64) (uint16, error) { return uint16(u), nil },
)

var NextNameDefault4 NextName = Next4default.ToNextName()
