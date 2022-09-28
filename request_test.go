package fsring

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestRequest(t *testing.T) {
	t.Parallel()

	var ITEST_FSRING_DIR string = os.Getenv("ITEST_FSRING_DIR")

	if len(ITEST_FSRING_DIR) < 1 {
		t.Skip("skipping tests using file system")
	}

	t.Run("NewHandler", func(t *testing.T) {
		t.Parallel()

		var chk NameChecker = NameCheckerNoCheck
		var root string = filepath.Join(ITEST_FSRING_DIR, "request/NewHandler")
		var n Next = NextBuilderNew(chk)(root)("manage.txt").
			Fallback(NextBuilderConst("0"), ErrIgnoreNew(fs.ErrNotExist))

		var wb WriteBuilder = WriteBuilder{}.Default()
		w, e := wb.BuildNoRename()
		mustNil(e)

		bf := WriteRequestHandlerBuilderFactory{
			Next:  n,
			Write: w,
			Dir:   root,
		}

		hb, e := bf.Build()
		mustNil(e)

		var h WriteRequestHandler = hb.NewHandler()

		ev, e := h(WriteRequest{data: []byte("hw")})
		mustNil(e)

		t.Run("dir check", check(ev.dir, root))
		t.Run("name check", check(ev.wroteName, filepath.Join(root, "0")))
	})
}
