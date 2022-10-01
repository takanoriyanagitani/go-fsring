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

	t.Run("Uint", func(t *testing.T) {
		t.Parallel()

		var wb WriteBuilder = WriteBuilder{}.Default()
		w, e := wb.BuildNoRename()
		mustNil(e)

		t.Run("write got", func(wtr Write) func(*testing.T) {
			return func(t *testing.T) {
				t.Parallel()

				t.Run("Uint8", func(t *testing.T) {
					t.Parallel()

					var root string = filepath.Join(ITEST_FSRING_DIR, "App/Uint/8")
					e := os.MkdirAll(root, 0755)
					mustNil(e)

					var head string = "head.txt"
					var tail string = "tail.txt"

					var chk NameChecker = NameCheckerNoCheck
					var emp IsEmpty = IsEmptyBuilderNew(chk)

					var mbf ManagerBuilderFactoryFs[uint8] = ManagerBuilderFactoryFs[uint8]{}.
						WithCheck(chk).
						WithGet(GetUintFsBuilderTxtHex3).
						WithName(filepath.Join(root, head)).
						WithSet(SetUintFsTxtHex3)
					hb, e := mbf.Build()
					mustNil(e)

					tb, e := mbf.
						WithName(filepath.Join(root, tail)).
						Build()
					mustNil(e)

					noent := func() (uint8, error) { return 0, nil }

					var mh ManagerUint[uint8] = hb.BuildManager().NoentIgnored(noent)
					var mt ManagerUint[uint8] = tb.BuildManager().NoentIgnored(noent)

					var rm RingMangerUint[uint8] = RingMangerUintNew(mh, mt, root)

					rhb, e := WriteRequestHandlerBuilderUintNew(
						rm,
						wtr.RejectNonEmpty(emp),
						uint2hex3,
					)
					mustNil(e)

					var wrh WriteRequestHandler = rhb.NewHandler()

					wehb, e := WroteEventHandlerBuilderUintNew(
						hex2uint3,
						rm,
					)
					mustNil(e)

					var weh WroteEventHandler = wehb.NewHandler()

					app, e := AppNew(wrh, weh)
					mustNil(e)

					t.Run("app got", func(a App) func(*testing.T) {
						return func(t *testing.T) {
							t.Parallel()

							t.Run("empty", func(t *testing.T) {
								t.Parallel()

								e := a.HandleWriteRequest(WriteRequestNew(nil))
								mustNil(e)
							})
						}
					}(app))
				})
			}
		}(w))
	})
}
