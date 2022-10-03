package fsring

import (
	"os"
	"syscall"
)

type FileSync func(f *os.File) error
type DirSync func(dir *os.File) error

var FileSyncAll FileSync = func(f *os.File) error { return f.Sync() }
var FileSyncData FileSync = func(f *os.File) error {
	return Compose(
		func(u uintptr) int { return int(u) },
		syscall.Fdatasync, // int -> error
	)(f.Fd())
}

var DirSyncDefault DirSync = func(f *os.File) error { return f.Sync() }

func DirnameSyncBuilderNew(chk NameChecker) func(DirSync) func(dirname string) error {
	return func(d DirSync) func(dirname string) error {
		return func(dirname string) error {
			f, e := os.Open(chk(dirname))
			if nil != e {
				return e
			}
			defer func() {
				_ = f.Close() // ignore close error after dir sync
			}()
			return d(f)
		}
	}
}
