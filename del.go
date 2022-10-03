package fsring

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type DeleteUint[T uint8 | uint16] func(target T) error

type DeleteRequest[T any] struct{ target T }

func DeleteRequestNew[T any](target T) DeleteRequest[T] { return DeleteRequest[T]{target} }

type RemovedEvent[T any] struct{ target T }

func (r RemovedEvent[T]) ToServiceEvent(e error) ServiceEvent {
	return OptUnwrapOrElse(
		func() (evt ServiceEvent, hasValue bool) {
			return OptFromBool(nil == e, Partial(ServiceEventOk, nil))
		},
		Partial(ServiceEventNg, e),
	)()
}

type RemovedEventHandler[T any] func(RemovedEvent[T]) error

type RemovedEventHandlerBuilderUint[T uint8 | uint16] struct {
	h2u hex2uint[T]
	u2h uint2hex[T]
	mng RingManagerUint[T]
}

func (b RemovedEventHandlerBuilderUint[T]) removed(removedName string) error {
	return b.mng.UpdateHead(b.h2u, filepath.Base(removedName))
}
func (b RemovedEventHandlerBuilderUint[T]) Removed(evt RemovedEvent[T]) error {
	var s string = b.u2h(evt.target)
	return b.removed(s)
}
func (b RemovedEventHandlerBuilderUint[T]) NewHandler() RemovedEventHandler[T] { return b.Removed }

func RemovedEventHandlerBuilderUintNew[T uint8 | uint16](
	h2u hex2uint[T],
	u2h uint2hex[T],
	mng RingManagerUint[T],
) (RemovedEventHandlerBuilderUint[T], error) {
	return ErrFromBool(
		nil != h2u && nil != u2h,
		func() RemovedEventHandlerBuilderUint[T] {
			return RemovedEventHandlerBuilderUint[T]{
				h2u,
				u2h,
				mng,
			}
		},
		func() error { return fmt.Errorf("Invalid arguments") },
	)
}

var RemovedEventHandlerBuilderUint3New func(
	RingManagerUint[uint8],
) (RemovedEventHandlerBuilderUint[uint8], error) = CurryErrIII(
	RemovedEventHandlerBuilderUintNew[uint8],
)(hex2uint3)(uint2hex3)

var RemovedEventHandlerBuilderUint4New func(
	RingManagerUint[uint16],
) (RemovedEventHandlerBuilderUint[uint16], error) = CurryErrIII(
	RemovedEventHandlerBuilderUintNew[uint16],
)(hex2uint4)(uint2hex4)

type DeleteHandler[T any] func(DeleteRequest[T]) (RemovedEvent[T], error)

func (d DeleteUint[T]) errIgnored(check func(error) (ignore bool)) DeleteUint[T] {
	return ErrIgnored(d, check)
}

func (d DeleteUint[T]) NewHandler() DeleteHandler[T] {
	return func(req DeleteRequest[T]) (RemovedEvent[T], error) {
		var tgt T = req.target
		var e error = d(tgt)
		return RemovedEvent[T]{target: tgt}, e
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
