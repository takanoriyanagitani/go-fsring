package fsring

import (
	"errors"
	"io"
	"io/fs"
	"os"
)

type Read func(filename string) (data []byte, e error)

type ReadByUint[T uint8 | uint16] func(key T) (data []byte, e error)

type ReadRequest[T any] struct{ target T }

func ReadRequestNew[T any](target T) ReadRequest[T] { return ReadRequest[T]{target} }

type ReadEvent struct{ data []byte }

func (r ReadEvent) ToServiceEvent(e error) ServiceEvent {
	return OptUnwrapOrElse(
		func() (evt ServiceEvent, hasValue bool) {
			return OptFromBool(
				nil == e,
				func() ServiceEvent {
					return ServiceEventOk(r.data).
						WithStatus(OptUnwrapOrElse(
							func() (status ServiceStatus, hasValue bool) {
								return OptFromBool(
									0 < len(r.data),
									func() ServiceStatus { return StatusOk },
								)
							},
							func() ServiceStatus { return StatusNotFound },
						)())
				},
			)
		},
		Partial(ServiceEventNg, e),
	)()
}

func (evt ReadEvent) WriteTo(w io.Writer) (int64, error) {
	return ComposeErr(
		w.Write,
		func(i int) (int64, error) { return int64(i), nil },
	)(evt.data)
}

type ReadHandler[T any] func(ReadRequest[T]) (ReadEvent, error)

func (r ReadEvent) Raw() []byte { return r.data }

func (r ReadByUint[T]) orElse(ef func(error) ([]byte, error)) ReadByUint[T] {
	return ErrOrElseGen(r, ef)
}

func (r ReadByUint[T]) NewHandler() ReadHandler[T] {
	return func(req ReadRequest[T]) (ReadEvent, error) {
		var tgt T = req.target
		data, e := r(tgt)
		return ReadEvent{data}, e
	}
}

func (r ReadByUint[T]) ErrorIgnored(ignoreMe error) ReadByUint[T] {
	ef := func(e error) ([]byte, error) {
		return ErrFromBool(
			errors.Is(e, ignoreMe),
			func() []byte { return nil },
			func() error { return e },
		)
	}
	return r.orElse(ef)
}

func (r ReadByUint[T]) NoentIgnored() ReadByUint[T] { return r.ErrorIgnored(fs.ErrNotExist) }

func ReadByUintBuilder[T uint8 | uint16](bld NameBuilderUint[T]) func(NameChecker) ReadByUint[T] {
	return func(chk NameChecker) ReadByUint[T] {
		var fullpath2data func(unchecked string) ([]byte, error) = ComposeErr(
			func(unchecked string) (checked string, e error) { return chk(unchecked), nil },
			os.ReadFile,
		)
		return ComposeErr(
			ErrFuncGen(bld), // T -> string, nil
			fullpath2data,   // string -> []byte, error
		)
	}
}
