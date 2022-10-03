package fsring

import (
	"testing"
)

func checkBuilder[T any](comp func(a, b T) (same bool)) func(got, expected T) func(*testing.T) {
	return func(got, expected T) func(*testing.T) {
		return func(t *testing.T) {
			var same bool = comp(got, expected)
			if !same {
				t.Errorf("Unexpected value got.\n")
				t.Errorf("Expected: %v\n", expected)
				t.Fatalf("Got:      %v\n", got)
			}
		}
	}
}

func check[T comparable](got, expected T) func(*testing.T) {
	return func(t *testing.T) {
		checkBuilder(func(a, b T) (same bool) { return a == b })(got, expected)(t)
	}
}

func mustNil(e error) {
	if nil != e {
		panic(e)
	}
}

func TestAll(t *testing.T) {
	t.Parallel()

	t.Run("RingServiceFactory", func(t *testing.T) {
		t.Parallel()

		t.Run("invalid factory", func(t *testing.T) {
			f := RingServiceFactory[uint8]{}
			_, e := f.Build()
			t.Run("Must fail", check(nil != e, true))
		})
	})
}
