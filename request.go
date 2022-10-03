package fsring

import (
	"fmt"
)

type WriteRequest struct{ data []byte }

func WriteRequestNew(data []byte) WriteRequest { return WriteRequest{data} }

type WriteRequestHandler func(WriteRequest) (WroteEvent, error)

type WriteRequestHandlerBuilder struct {
	nex Next
	wtr Write
	dir string
}

type WriteRequestHandlerBuilderFactory struct {
	Next
	Write
	Dir string
}

func (f WriteRequestHandlerBuilderFactory) Build() (WriteRequestHandlerBuilder, error) {
	var valid bool = IterFromArr([]bool{
		nil != f.Next,
		nil != f.Write,
		0 < len(f.Dir),
	}).All(Identity[bool])
	return ErrFromBool(
		valid,
		func() WriteRequestHandlerBuilder {
			return WriteRequestHandlerBuilder{
				nex: f.Next,
				wtr: f.Write,
				dir: f.Dir,
			}
		},
		func() error { return fmt.Errorf("Invalid factory") },
	)
}

func (m WriteRequestHandlerBuilder) Write2Next(data []byte) (ev WroteEvent, e error) {
	var dir2next func(dir string) (next string, e error) = ComposeErr(
		m.nex,
		func(next string) (string, error) {
			return ComposeErr(
				func(_ string) (int, error) { return m.wtr(next, data) },
				func(_ int) (string, error) { return next, nil },
			)(next)
		},
	)
	next, e := dir2next(m.dir)
	return WroteEvent{dir: m.dir, wroteName: next}, e
}

func (m WriteRequestHandlerBuilder) NewHandler() WriteRequestHandler {
	return func(req WriteRequest) (evt WroteEvent, e error) { return m.Write2Next(req.data) }
}

type WriteRequestHandlerBuilderUint[T uint8 | uint16] struct {
	mng RingManagerUint[T]
	wtr Write
	cnv uint2hex[T]
}

func (b WriteRequestHandlerBuilderUint[T]) nextPath() (string, error) {
	return b.mng.nextPath(b.cnv)
}

func (b WriteRequestHandlerBuilderUint[T]) Write2Next(data []byte) (wroteName string, e error) {
	var wdata func(string) func([]byte) (int, error) = CurryErr(b.wtr)
	return ComposeErr(
		func(m RingManagerUint[T]) (string, error) { return m.nextPath(b.cnv) },
		func(name string) (string, error) {
			return ComposeErr(
				wdata(name), // []byte -> int, error
				func(_ int) (string, error) { return name, nil },
			)(data)
		},
	)(b.mng)
}

func (b WriteRequestHandlerBuilderUint[T]) Write(req WriteRequest) (WroteEvent, error) {
	return ComposeErr(
		b.Write2Next, // []byte -> string, error
		func(wroteName string) (WroteEvent, error) {
			var dir string = b.mng.dir
			return WroteEvent{
				dir,
				wroteName,
			}, nil
		},
	)(req.data)
}

func (b WriteRequestHandlerBuilderUint[T]) NewHandler() WriteRequestHandler { return b.Write }

func WriteRequestHandlerBuilderUintNew[T uint8 | uint16](
	mng RingManagerUint[T],
	wtr Write,
	cnv uint2hex[T],
) (WriteRequestHandlerBuilderUint[T], error) {
	var valid bool = nil != wtr && nil != cnv
	return ErrFromBool(
		valid,
		func() WriteRequestHandlerBuilderUint[T] {
			return WriteRequestHandlerBuilderUint[T]{
				mng,
				wtr,
				cnv,
			}
		},
		func() error { return fmt.Errorf("Invalid arguments") },
	)
}
