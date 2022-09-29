package fsring

import (
	"fmt"
	"io"
	"os"
)

type GetUint[T uint8 | uint16] func() (T, error)
type SetUint[T uint8 | uint16] func(neo T) error

func getUintNewTxt[T uint8 | uint16](conv func(string) (T, error)) func(string) GetUint[T] {
	return func(s string) GetUint[T] {
		return func() (T, error) {
			return ComposeErr(
				conv,
				ErrFuncGen(Identity[T]),
			)(s)
		}
	}
}

var getUintNewTxtHex3 func(string) GetUint[uint8] = getUintNewTxt(hex2uint3)
var getUintNewTxtHex4 func(string) GetUint[uint16] = getUintNewTxt(hex2uint4)

type GetUintFs[T uint8 | uint16] func(NameChecker) func(filename string) GetUint[T]

func GetUintFsBuilderTxt[T uint8 | uint16](s2g func(string) GetUint[T]) GetUintFs[T] {
	return func(chk NameChecker) func(string) GetUint[T] {
		return func(filename string) GetUint[T] {
			return func() (T, error) {
				return ComposeErr(
					os.ReadFile, // string -> []byte, error
					func(b []byte) (T, error) { return s2g(string(b))() },
				)(chk(filename))
			}
		}
	}
}

var GetUintFsBuilderTxtHex3 GetUintFs[uint8] = GetUintFsBuilderTxt(getUintNewTxtHex3)
var GetUintFsBuilderTxtHex4 GetUintFs[uint16] = GetUintFsBuilderTxt(getUintNewTxtHex4)

func uint2hexBuilder[T uint8 | uint16](format string) func(T) string {
	return func(t T) string { return fmt.Sprintf(format, t) }
}

var uint2hex3 func(uint8) string = uint2hexBuilder[uint8]("%02x")
var uint2hex4 func(uint16) string = uint2hexBuilder[uint16]("%04x")

func setUintWriteNew[T uint8 | uint16](conv func(T) string) func(io.Writer) SetUint[T] {
	return func(w io.Writer) SetUint[T] {
		f := ComposeErr(
			ErrFuncGen(conv), // T -> string
			func(s string) (int, error) { return w.Write([]byte(s)) },
		)
		return ErrOnly(f)
	}
}

var setUintWriterTxtHex3 func(io.Writer) SetUint[uint8] = setUintWriteNew(uint2hex3)
var setUintWriterTxtHex4 func(io.Writer) SetUint[uint16] = setUintWriteNew(uint2hex4)

type SetUintFs[T uint8 | uint16] func(NameChecker) func(filename string) SetUint[T]

func SetUintFsBuilder[T uint8 | uint16](w2s func(io.Writer) SetUint[T]) SetUintFs[T] {
	return func(chk NameChecker) func(string) SetUint[T] {
		return func(filename string) SetUint[T] {
			return func(neo T) error {
				f, e := os.Create(chk(filename))
				if nil != e {
					return e
				}
				defer func() {
					_ = f.Close() // ignore close error after sync
				}()
				var su SetUint[T] = w2s(f)
				return Err1st([]func() error{
					func() error { return su(neo) },
					func() error { return f.Sync() },
				})
			}
		}
	}
}

var SetUintFsTxtHex3 SetUintFs[uint8] = SetUintFsBuilder(setUintWriterTxtHex3)
var SetUintFsTxtHex4 SetUintFs[uint16] = SetUintFsBuilder(setUintWriterTxtHex4)
