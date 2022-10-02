package fsring

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type DeleteUint[T uint8 | uint16] func(target T) error

type DeleteRequest[T any] struct{ target T }

func DeleteRequestNew[T any](target T) DeleteRequest[T] { return DeleteRequest[T]{target} }

type RemovedEvent struct{}

func (r RemovedEvent) WriteTo(_ io.Writer) (int64, error) { return 0, nil }

func (r RemovedEvent) ToServiceEvent(e error) ServiceEvent {
	return OptUnwrapOrElse(
		func() (evt ServiceEvent, hasValue bool) {
			return OptFromBool(nil == e, Partial(ServiceEventOk, nil))
		},
		Partial(ServiceEventNg, e),
	)()
}

type DeleteHandler[T any] func(DeleteRequest[T]) (RemovedEvent, error)

func (d DeleteUint[T]) errIgnored(check func(error) (ignore bool)) DeleteUint[T] {
	return ErrIgnored(d, check)
}

func (d DeleteUint[T]) NewHandler() DeleteHandler[T] {
	return func(req DeleteRequest[T]) (RemovedEvent, error) {
		var tgt T = req.target
		var e error = d(tgt)
		return RemovedEvent{}, e
	}
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
