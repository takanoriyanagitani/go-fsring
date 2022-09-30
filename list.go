package fsring

type List func(dirname string) (filenames []string, e error)

type ListUint[T uint8 | uint16] func(dirname string) (names []T, e error)

func count2listUint[T uint8 | uint16](cnt int, lbi T) Iter[T] {
	var i Iter[int] = IterInts(0, cnt)
	var t T = lbi
	return IterMap(i, func(_ int) T {
		var ret T = t
		t += 1
		return ret
	})
}
