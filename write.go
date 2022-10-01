package fsring

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var ErrNonEmpty error = errors.New("write request to non-empty file rejected")

type Write func(filename string, data []byte) (wrote int, e error)

// RejectNonEmpty creates new Write which rejects writes to non-empty file.
// Error must be ErrNonEmpty if the file was not empty.
func (w Write) RejectNonEmpty(ie IsEmpty) Write {
	return func(mayBeNonEmpty string, data []byte) (int, error) {
		var wtr func([]byte) (int, error) = CurryErr(w)(mayBeNonEmpty)
		var name2dat func(string) ([]byte, error) = ComposeErr(
			ie, // string -> empty bool, e error
			func(empty bool) ([]byte, error) {
				return ErrFromBool(
					empty,
					func() []byte { return data },
					func() error { return ErrNonEmpty },
				)
			},
		)
		return ComposeErr(
			name2dat,
			wtr,
		)(mayBeNonEmpty)
	}
}

type NameChecker func(unchecked string) (checked string)

func (chk NameChecker) NewFileCreate(mode fs.FileMode) func(filename string) (*os.File, error) {
	f := func(filename string) (*os.File, error) {
		var dirname string = filepath.Dir(filename)
		return ComposeErr(
			func(_ string) (string, error) { return filename, os.MkdirAll(dirname, mode) },
			os.Create,
		)(dirname)
	}
	return ComposeErr(
		func(unchecked string) (checked string, e error) { return chk(unchecked), nil },
		f,
	)
}

var NameCheckerNoCheck NameChecker = Identity[string]

var NameCheckerNoEmpty NameChecker = func(unchecked string) (checked string) {
	if len(unchecked) < 1 {
		panic("Invalid name")
	}
	return unchecked
}

// WriteBuilderNewNoRename creates Write.
// Data must be self-checked(use gzip, bzip2, ...)
func WriteBuilderNewNoRename(chk NameChecker) func(FileSync) func(DirSync) Write {
	return func(filesync FileSync) func(DirSync) Write {
		return func(d DirSync) Write {
			var syncDir func(dirname string) error = DirnameSyncBuilderNew(chk)(d)
			var filename2dir2sync func(filename string) error = Compose(
				filepath.Dir,
				syncDir,
			)
			return func(filename string, data []byte) (wrote int, e error) {
				f, e := chk.NewFileCreate(0755)(filename)
				if nil != e {
					return -1, fmt.Errorf("Unable to create: %v(%s)", e, filename)
				}
				defer func() {
					_ = f.Close() // ignore close error after fsync/fdatasync
				}()

				wrote, e = f.Write(data)
				e = Err1st([]func() error{
					func() error { return e },
					func() error { return filesync(f) },
					func() error { return filename2dir2sync(filename) },
				})
				return
			}
		}
	}
}

type WriteBuilder struct {
	NameChecker
	FileSync
	DirSync
}

func (b WriteBuilder) Default() WriteBuilder {
	b.NameChecker = NameCheckerNoCheck
	b.FileSync = FileSyncAll
	b.DirSync = DirSyncDefault
	return b
}

func (b WriteBuilder) BuildNoRename() (Write, error) {
	var valid bool = IterFromArr([]bool{
		nil != b.NameChecker,
		nil != b.FileSync,
		nil != b.DirSync,
	}).All(Identity[bool])
	return ErrFromBool(
		valid,
		func() Write { return WriteBuilderNewNoRename(b.NameChecker)(b.FileSync)(b.DirSync) },
		func() error { return fmt.Errorf("Invalid builder") },
	)
}
