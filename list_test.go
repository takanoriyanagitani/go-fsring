package fsring

import (
	"testing"
)

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("count2listUint", func(t *testing.T) {
		t.Parallel()

		t.Run("zero", func(t *testing.T) {
			t.Parallel()

			var i Iter[uint8] = count2listUint[uint8](0, 42)
			_, hasValue := i()
			t.Run("Must be empty", check(hasValue, false))
		})

		t.Run("single item", func(t *testing.T) {
			t.Parallel()

			var i Iter[uint8] = count2listUint[uint8](1, 42)
			u, hasValue := i()
			t.Run("Must not be empty", check(hasValue, true))
			t.Run("Must be same", check(u, 42))
		})

		t.Run("many integers", func(t *testing.T) {
			t.Parallel()

			var i Iter[uint8] = count2listUint[uint8](3, 42)
			var tot uint8 = i.Reduce(0, func(a, b uint8) uint8 { return a + b })
			t.Run("Must be same", check(tot, 42+43+44))
		})

		t.Run("wraparound", func(t *testing.T) {
			t.Parallel()

			var i Iter[uint8] = count2listUint[uint8](3, 254)
			var tot uint8 = i.Reduce(0, func(a, b uint8) uint8 { return a + b })
			t.Run("Must be same", check(tot, 253)) // 254+255+0 = 254-1+256 = 253
		})
	})
}
