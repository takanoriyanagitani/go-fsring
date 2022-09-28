package fsring

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestNext(t *testing.T) {
	t.Parallel()

	t.Run("Next4default", func(t *testing.T) {
		t.Parallel()

		t.Run("zero", func(t *testing.T) {
			t.Parallel()

			var n NextName4 = Next4default

			nex, e := n(0)
			mustNil(e)

			t.Run("Must be same", check(nex, 1))
		})

		t.Run("max", func(t *testing.T) {
			t.Parallel()

			var n NextName4 = Next4default

			nex, e := n(65535)
			mustNil(e)

			t.Run("Must be same", check(nex, 0))
		})
	})

	t.Run("NextNameDefault4", func(t *testing.T) {
		t.Parallel()

		var n NextName = NextNameDefault4

		t.Run("zero", func(t *testing.T) {
			t.Parallel()
			nex, e := n("0000")
			mustNil(e)
			t.Run("Must be same", check(nex, "0001"))
		})

		t.Run("max", func(t *testing.T) {
			t.Parallel()
			nex, e := n("ffff")
			mustNil(e)
			t.Run("Must be same", check(nex, "0000"))
		})

	})

	t.Run("filesystem test", func(t *testing.T) {
		t.Parallel()

		var ITEST_FSRING_DIR string = os.Getenv("ITEST_FSRING_DIR")
		if len(ITEST_FSRING_DIR) < 1 {
			t.Skip("skipping tests using file system")
		}

		t.Run("NextBuilderNew", func(t *testing.T) {
			t.Parallel()

			var root string = filepath.Join(ITEST_FSRING_DIR, "NextBuilderNew")

			e := os.MkdirAll(root, 0755)
			mustNil(e)

			var chk NameChecker = NameCheckerNoCheck

			var n Next = NextBuilderNew(chk)(root)("manage.txt").
				Fallback(NextBuilderConst("0"), ErrIgnoreNew(fs.ErrNotExist))
			nex, e := n(root)
			mustNil(e)
			t.Run("Must be same", check(nex, filepath.Join(root, "0")))
		})
	})
}
