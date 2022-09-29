package fsring

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type WroteEvent struct {
	dir       string
	wroteName string
}

func (we WroteEvent) ToNext(n NextName) (next string, e error) { return n(we.wroteName) }
func (we WroteEvent) Next2Writer(n NextName, w io.Writer) (wrote int, e error) {
	return ComposeErr(
		func(_ NextName) (next string, e error) { return we.ToNext(n) },
		func(next string) (int, error) { return w.Write([]byte(next)) },
	)(n)
}

type WroteEventHandler func(WroteEvent) error
type WroteEventHandlerBuilder func(NextName) WroteEventHandler

func WroteEventHandlerBuilderManaged(chk NameChecker) func(mname string) WroteEventHandlerBuilder {
	return func(managename string) WroteEventHandlerBuilder {
		return func(nex NextName) WroteEventHandler {
			return func(evt WroteEvent) error {
				var toWriter func(io.Writer) (int, error) = CurryErr(evt.Next2Writer)(nex)

				var fullmng string = filepath.Join(evt.dir, managename)
				f, e := os.Create(chk(fullmng))
				if nil != e {
					return fmt.Errorf("Unable to create manage txt: %v", e)
				}
				defer func() {
					_ = f.Close() // ignore close error after fsync
				}()

				return Err1st([]func() error{
					func() error { return ErrOnly(toWriter)(f) },
					func() error { return f.Sync() },
				})
			}
		}
	}
}

type WroteEventHandlerBuilderUint[T uint8 | uint16] struct {
	h2u hex2uint[T]
	mng RingMangerUint[T]
}

func (b WroteEventHandlerBuilderUint[T]) write(wroteName string) error {
	return b.mng.UpdateTail(b.h2u, wroteName)
}

func (b WroteEventHandlerBuilderUint[T]) Write(evt WroteEvent) error {
	return b.write(evt.wroteName)
}

func (b WroteEventHandlerBuilderUint[T]) NewHandler() WroteEventHandler { return b.Write }

func WroteEventHandlerBuilderUintNew[T uint8 | uint16](
	h2u hex2uint[T],
	mng RingMangerUint[T],
) (WroteEventHandlerBuilderUint[T], error) {
	return ErrFromBool(
		nil != h2u,
		func() WroteEventHandlerBuilderUint[T] {
			return WroteEventHandlerBuilderUint[T]{
				h2u,
				mng,
			}
		},
		func() error { return fmt.Errorf("Invalid arguments") },
	)
}
