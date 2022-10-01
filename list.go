package fsring

type List func(dirname string) (filenames []string, e error)

type ListUint[T uint8 | uint16] func() (names Iter[T], e error)

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
