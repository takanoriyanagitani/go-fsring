package fsring

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestApp(t *testing.T) {
	t.Parallel()

	var ITEST_FSRING_DIR string = os.Getenv("ITEST_FSRING_DIR")

	if len(ITEST_FSRING_DIR) < 1 {
		t.Skip("skipping tests using filesystem")
	}

	t.Run("HandleWriteRequest", func(t *testing.T) {
		t.Parallel()

		var root string = filepath.Join(ITEST_FSRING_DIR, "App/HandleWriteRequest")

		var mng string = "mng.txt"

		var chk NameChecker = NameCheckerNoEmpty
		var n Next = NextBuilderNew(chk)(root)(mng).
			Fallback(NextBuilderConst("0000"), ErrIgnoreNew(fs.ErrNotExist))

		var wb WriteBuilder = WriteBuilder{}.Default()
		w, e := wb.BuildNoRename()
		mustNil(e)

		rhbf := WriteRequestHandlerBuilderFactory{
			Next:  n,
			Write: w,
			Dir:   root,
		}

		rhb, e := rhbf.Build()
		mustNil(e)

		var rh WriteRequestHandler = rhb.NewHandler()

		var ngen NextName = NextNameDefault4

		var ehb WroteEventHandlerBuilder = WroteEventHandlerBuilderManaged(chk)(mng)
		var eh WroteEventHandler = ehb(ngen)

		app := App{
			wreqh: rh,
			wevth: eh,
		}

		e = app.HandleWriteRequest(WriteRequest{
			data: []byte("hw"),
		})
		mustNil(e)
	})
}
