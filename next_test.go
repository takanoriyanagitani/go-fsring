package fsring

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestNext(t *testing.T) {
	t.Parallel()

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

	t.Run("NextNameDefault3", func(t *testing.T) {
		t.Parallel()

		var n NextName = NextNameDefault3

		t.Run("zero", func(t *testing.T) {
			t.Parallel()
			nex, e := n("00")
			mustNil(e)
			t.Run("Must be same", check(nex, "01"))
		})

		t.Run("max", func(t *testing.T) {
			t.Parallel()
			nex, e := n("ff")
			mustNil(e)
			t.Run("Must be same", check(nex, "00"))
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

			var root string = filepath.Join(ITEST_FSRING_DIR, "Next/NextBuilderNew")

			e := os.MkdirAll(root, 0755)
			mustNil(e)

			var chk NameChecker = NameCheckerNoCheck

			var n Next = NextBuilderNew(chk)(root)("manage.txt").
				Fallback(NextBuilderConst("0"), ErrIgnoreNew(fs.ErrNotExist))
			nex, e := n(root)
			mustNil(e)
			t.Run("Must be same", check(nex, filepath.Join(root, "0")))
		})

		t.Run("FallbackIfNotEmpty", func(t *testing.T) {
			t.Parallel()

			e := os.RemoveAll(filepath.Join(ITEST_FSRING_DIR, "Next/FallbackIfNotEmpty"))
			mustNil(e)

			t.Run("NextBuilderHeavy4", func(t *testing.T) {
				t.Parallel()
				var root string = filepath.Join(ITEST_FSRING_DIR, "Next/FallbackIfNotEmpty/16")

				e := os.MkdirAll(root, 0755)
				mustNil(e)

				var chk NameChecker = NameCheckerNoCheck
				var ie IsEmpty = IsEmptyBuilderNew(chk)

				var mng string = "manage.txt"
				var wrongNext string = filepath.Join(root, "0000")

				e = os.WriteFile(filepath.Join(root, mng), []byte(wrongNext), 0755)
				mustNil(e)

				var n Next = NextBuilderNew(chk)(root)(mng).
					FallbackIfNotEmpty(ie, NextBuilderHeavy4(ie))

				f, e := os.Create(wrongNext)
				mustNil(e)
				_, e = f.Write([]byte("not empty"))
				mustNil(e)
				f.Close()

				nex, e := n(root)
				mustNil(e)
				t.Run("Must be same", check(nex, filepath.Join(root, "0001")))
			})

			t.Run("NextBuilderHeavy3", func(t *testing.T) {
				t.Parallel()
				var root string = filepath.Join(ITEST_FSRING_DIR, "Next/FallbackIfNotEmpty/8")

				e := os.MkdirAll(root, 0755)
				mustNil(e)

				var chk NameChecker = NameCheckerNoCheck
				var ie IsEmpty = IsEmptyBuilderNew(chk)

				var mng string = "manage.txt"
				var wrongNext string = filepath.Join(root, "00")

				e = os.WriteFile(filepath.Join(root, mng), []byte(wrongNext), 0755)
				mustNil(e)

				var n Next = NextBuilderNew(chk)(root)(mng).
					FallbackIfNotEmpty(ie, NextBuilderHeavy3(ie))

				f, e := os.Create(wrongNext)
				mustNil(e)
				_, e = f.Write([]byte("not empty"))
				mustNil(e)
				f.Close()

				nex, e := n(root)
				mustNil(e)
				t.Run("Must be same", check(nex, filepath.Join(root, "01")))

				for i := 0; i < 256; i++ {
					name := filepath.Join(root, fmt.Sprintf("%02x", i))
					f, e := os.Create(name)
					mustNil(e)
					defer f.Close()
					_, e = f.Write([]byte("not empty"))
					mustNil(e)
				}

				_, e = n(root)
				t.Run("Must fail(too many queue)", check(nil != e, true))
				t.Run("Must same(too many queue)", check(errors.Is(e, ErrTooManyQueue), true))
			})

		})
	})
}
