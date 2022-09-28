package fsring

// ArrayFilter creates filtered array.
// Elements which does not satisfy provided condition will disappear.
func ArrayFilter[T any](f func(T) (ok bool), a []T) (b []T) {
	for _, item := range a {
		if f(item) {
			b = append(b, item)
		}
	}
	return b
}
