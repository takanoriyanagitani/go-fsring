package fsring

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDel(t *testing.T) {
	t.Parallel()

	t.Run("DeleteUint", func(t *testing.T) {
		t.Parallel()

		var ITEST_FSRING_DIR string = os.Getenv("ITEST_FSRING_DIR")
		if len(ITEST_FSRING_DIR) < 1 {
			t.Skip("skipping tests using filesystem")
		}

		var root string = filepath.Join(ITEST_FSRING_DIR, "DeleteUint")

		t.Run("NameBuilderUint3", func(t *testing.T) {
			t.Parallel()

			var dir string = filepath.Join(root, "NameBuilderUint3")

			var nbu NameBuilderUint[uint8] = NameBuilderUint3(dir)
			var dub func(NameChecker) DeleteUint[uint8] = DeleteUintBuilder(nbu)
			var nchk NameChecker = NameCheckerNoCheck

			var du DeleteUint[uint8] = dub(nchk).
				NoentIgnored()

			t.Run("not exist", func(t *testing.T) {
				t.Parallel()

				e := du(0x42)
				mustNil(e)
			})

			t.Run("empty", func(t *testing.T) {
				t.Parallel()

				e := os.MkdirAll(dir, 0755)
				mustNil(e)
				e = os.WriteFile(filepath.Join(dir, "3f"), nil, 0644)
				mustNil(e)

				e = du(0x3f)
				mustNil(e)

				s, e := os.Stat(filepath.Join(dir, "3f"))
				mustNil(e)
				t.Run("Must same", check(s.Size(), 0))
			})

			t.Run("non empty", func(t *testing.T) {
				t.Parallel()

				e := os.MkdirAll(dir, 0755)
				mustNil(e)
				e = os.WriteFile(filepath.Join(dir, "ff"), []byte("hw"), 0644)
				mustNil(e)

				e = du(0xff)
				mustNil(e)

				s, e := os.Stat(filepath.Join(dir, "ff"))
				mustNil(e)
				t.Run("Must same", check(s.Size(), 0))
			})
		})
	})
}
