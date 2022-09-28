package fsring

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type IsEmpty func(filename string) (empty bool, e error)

func IsEmptyBuilderNew(chk NameChecker) IsEmpty {
	return func(filename string) (empty bool, e error) {
		f, e := os.Open(chk(filename))
		if nil != e {
			if errors.Is(e, fs.ErrNotExist) {
				return true, nil
			}
			return false, fmt.Errorf("Unexpected error: %v", e)
		}
		defer func() {
			_ = f.Close() // ignore close error after file check
		}()
		return isEmptyFile(f)
	}
}

var isEmptyFile func(fs.File) (empty bool, e error) = ComposeErr(
	func(file fs.File) (fs.FileInfo, error) { return file.Stat() },
	isEmptyStat,
)

func isEmptyStat(f fs.FileInfo) (empty bool, e error) { return 0 == f.Size(), nil }
