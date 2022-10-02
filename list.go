package fsring

type List func(dirname string) (filenames []string, e error)

type ListUint[T uint8 | uint16] func() (names Iter[T], e error)

type ListRequest struct{}

type ListEvent[T any] struct{ basenames []T }

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
