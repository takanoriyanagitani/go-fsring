package fsring

import (
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
