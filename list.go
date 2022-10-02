package fsring

import (
	"bytes"
	"fmt"
	"io"
)

type List func(dirname string) (filenames []string, e error)

type ListUint[T uint8 | uint16] func() (names Iter[T], e error)

type ListRequest struct{}

type ListEvent[T any] struct{ basenames []T }

type ListEventWriterTo[T any] func(io.Writer) func(ListEvent[T]) (int64, error)

func (l ListEvent[T]) ToServiceEvent(e error, wtr ListEventWriterTo[T]) ServiceEvent {
	return OptUnwrapOrElse(
		func() (evt ServiceEvent, hasValue bool) {
			return OptFromBool(
				nil == e,
				func() ServiceEvent {
					var buf bytes.Buffer
					_, e := wtr(&buf)(l)
					return ServiceEventNew(buf.Bytes(), e)
				},
			)
		},
		Partial(ServiceEventNg, e),
	)()
}

func bytes2writer(w io.Writer) func([]byte) (int64, error) {
	return ComposeErr(
		w.Write, // []byte -> int, error
		func(i int) (int64, error) { return int64(i), nil },
	)
}

type Uint2Writer[T uint8 | uint16] func(io.Writer) func(T) (int64, error)

func (u Uint2Writer[T]) NewEventWriter() ListEventWriterTo[T] {
	return func(w io.Writer) func(ListEvent[T]) (int64, error) {
		var t2wtr func(T) (int64, error) = u(w)
		return func(evt ListEvent[T]) (int64, error) {
			var names Iter[T] = IterFromArr(evt.basenames)
			return IterTryFold(
				names,
				0,
				func(state int64, item T) (int64, error) {
					return ComposeErr(
						t2wtr, // T -> int64, error
						func(cnt int64) (int64, error) { return state + cnt, nil },
					)(item)
				},
			)
		}
	}
}

func (u Uint2Writer[T]) WithSuffix(suffix []byte) Uint2Writer[T] {
	return func(w io.Writer) func(T) (int64, error) {
		return ComposeErr(
			u(w), // T -> int64, error
			func(c1 int64) (int64, error) {
				return ComposeErr(
					bytes2writer(w), // []byte -> int64, error
					func(c2 int64) (int64, error) { return c1 + c2, nil },
				)(suffix)
			},
		)
	}
}

func Uint2WriterHexTxtNew[T uint8 | uint16](format string) Uint2Writer[T] {
	writeUint := func(w io.Writer) func(T) (int, error) {
		return func(u T) (int, error) {
			return fmt.Fprintf(w, format, u)
		}
	}
	return func(w io.Writer) func(T) (int64, error) {
		var t2wtr func(T) (int, error) = writeUint(w)
		return ComposeErr(
			t2wtr,
			func(i int) (int64, error) { return int64(i), nil },
		)
	}
}

var Uint2WriterHexTxt3 Uint2Writer[uint8] = Uint2WriterHexTxtNew[uint8]("%02x")
var Uint2WriterHexTxt4 Uint2Writer[uint16] = Uint2WriterHexTxtNew[uint16]("%04x")

func (l ListEvent[T]) BaseNames() []T { return l.basenames }

type ListRequestHandler[T any] func(ListRequest) (ListEvent[T], error)

func (l ListUint[T]) Fallback(other ListUint[T]) ListUint[T] {
	f := ErrFallback(
		IgnoreArg[int](l),
		IgnoreArg[int](other),
	)
	return func() (Iter[T], error) { return f(0) }
}

func (l ListUint[T]) NewHandler() ListRequestHandler[T] {
	return func(_ ListRequest) (evt ListEvent[T], e error) {
		return ComposeErr(
			func(_ ListUint[T]) (names Iter[T], e error) { return l() },
			func(names Iter[T]) (ListEvent[T], error) {
				var basenames []T = names.ToArray()
				return ListEvent[T]{basenames}, nil
			},
		)(l)
	}
}

func ListUintFallbackNew[T uint8 | uint16](max int) ListUint[T] {
	return func() (names Iter[T], e error) {
		var u T = 0
		var i Iter[int] = IterInts(0, max)
		return IterMap(i, func(_ int) T {
			var ret T = u
			u += 1
			return ret
		}), nil
	}
}

var ListUintFallbackNew3 ListUint[uint8] = ListUintFallbackNew[uint8](256)
var ListUintFallbackNew4 ListUint[uint16] = ListUintFallbackNew[uint16](65536)

func count2listUint[T uint8 | uint16](cnt int, lbi T) Iter[T] {
	var i Iter[int] = IterInts(0, cnt)
	var t T = lbi
	return IterMap(i, func(_ int) T {
		var ret T = t
		t += 1
		return ret
	})
}

func subNew[T uint8 | uint16 | uint32 | uint64](i int) func(x, y, borrow T) (diff, borrowOut T) {
	return func(x, y, borrow T) (diff, borrowOut T) {
		diff = x - y - borrow
		borrowOut = ((^x & y) | (^(x ^ y) & diff)) >> i
		return
	}
}

type sub[T uint8 | uint16 | uint32 | uint64] func(x, y, borrow T) (diff, borrowOut T)

var sub3 sub[uint8] = subNew[uint8](7)
var sub4 sub[uint16] = subNew[uint16](15)

func htcountNew[T uint8 | uint16](s sub[T]) func(tail, head T) T {
	return func(tail, head T) T {
		diff, _ := s(tail, head, 0)
		return diff
	}
}

type HeadTailCounter[T uint8 | uint16] func(tail, head T) T

var HeadTailCounter3 HeadTailCounter[uint8] = htcountNew(sub3)
var HeadTailCounter4 HeadTailCounter[uint16] = htcountNew(sub4)
