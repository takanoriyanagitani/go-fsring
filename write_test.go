package fsring

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWrite(t *testing.T) {
	t.Parallel()

	t.Run("RejectNonEmpty", func(t *testing.T) {
		t.Parallel()

		var wb WriteBuilder = WriteBuilder{}.Default()
		w, e := wb.BuildNoRename()
		mustNil(e)

		var chk NameChecker = NameCheckerNoCheck
		var emp IsEmpty = IsEmptyBuilderNew(chk)

		t.Run("writer got", func(wtr Write) func(*testing.T) {
			return func(t *testing.T) {
				t.Parallel()

				var ITEST_FSRING_DIR string = os.Getenv("ITEST_FSRING_DIR")
				if len(ITEST_FSRING_DIR) < 1 {
					t.Skip("skipping tests using filesystem")
				}

				var root string = filepath.Join(ITEST_FSRING_DIR, "Write/RejectNonEmpty")

				t.Run("noent", func(t *testing.T) {
					t.Parallel()

					var dir string = filepath.Join(root, "noent")
					e := os.RemoveAll(dir)
					mustNil(e)
					e = os.MkdirAll(dir, 0755)
					mustNil(e)

					_, e = wtr("empty.dat", nil)
					mustNil(e)
				})

				t.Run("empty", func(t *testing.T) {
					t.Parallel()

					var dir string = filepath.Join(root, "empty")
					e := os.RemoveAll(dir)
					mustNil(e)
					e = os.MkdirAll(dir, 0755)
					mustNil(e)

					var name string = filepath.Join(dir, "empty.dat")

					e = os.WriteFile(name, nil, 0644)
					mustNil(e)

					_, e = wtr(name, nil)
					mustNil(e)
				})

				t.Run("non empty", func(t *testing.T) {
					t.Parallel()

					var dir string = filepath.Join(root, "non-empty")
					e := os.RemoveAll(dir)
					mustNil(e)
					e = os.MkdirAll(dir, 0755)
					mustNil(e)

					var name string = filepath.Join(dir, "non-empty.dat")

					e = os.WriteFile(name, []byte("do not overwrite me"), 0644)
					mustNil(e)

					_, e = wtr(name, []byte("ignore me"))
					t.Run("Must fail", check(nil != e, true))
				})
			}
		}(w.RejectNonEmpty(emp)))
	})
}
