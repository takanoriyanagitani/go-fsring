package fsring

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRead(t *testing.T) {
	t.Parallel()

	t.Run("ReadByUintBuilder", func(t *testing.T) {
		t.Parallel()

		var ITEST_FSRING_DIR string = os.Getenv("ITEST_FSRING_DIR")
		if len(ITEST_FSRING_DIR) < 1 {
			t.Skip("skipping tests using filesystem")
		}

		var root string = filepath.Join(ITEST_FSRING_DIR, "Read/ReadByUintBuilder")

		t.Run("uint8", func(t *testing.T) {
			t.Parallel()

			var nb3 func(dirname string) NameBuilderUint[uint8] = NameBuilderUint3
			var dir string = filepath.Join(root, "uint8")

			e := os.MkdirAll(dir, 0755)
			mustNil(e)

			var nb NameBuilderUint[uint8] = nb3(dir)
			var chk NameChecker = NameCheckerNoCheck

			var rbub ReadByUint[uint8] = ReadByUintBuilder(nb)(chk).
				NoentIgnored()

			t.Run("ReadByUint got", func(ru ReadByUint[uint8]) func(*testing.T) {
				return func(t *testing.T) {
					t.Parallel()

					const missing uint8 = 0x37
					const empty uint8 = 0x42

					t.Run("missing file", func(t *testing.T) {
						t.Parallel()

						dat, e := ru(missing)
						mustNil(e)
						t.Run("Must same", check(len(dat), 0))
					})

					t.Run("empty file", func(t *testing.T) {
						t.Parallel()

						e := os.WriteFile(filepath.Join(dir, "42"), nil, 0644)
						mustNil(e)

						dat, e := ru(empty)
						mustNil(e)
						t.Run("Must same", check(len(dat), 0))
					})
				}
			}(rbub))
		})
	})
}
