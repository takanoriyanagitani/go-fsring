package fsring

import (
	"fmt"
)

type WriteRequest struct{ data []byte }

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
