package fsring

import (
	"testing"
	"os"
	"path/filepath"
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
			t.Parallel()

			f := RingServiceFactory[uint8]{}
			_, e := f.Build()
			t.Run("Must fail", check(nil != e, true))
		})

		t.Run("valid factory", func(t *testing.T){
			t.Parallel()

			var ITEST_FSRING_DIR string = os.Getenv("ITEST_FSRING_DIR")
			if len(ITEST_FSRING_DIR) < 1 {
				t.Skip("skipping tests using file system")
			}

			var root string = filepath.Join(ITEST_FSRING_DIR, "fsring/valid/factory")

			var guf GetUintFs[uint8] = GetUintFsBuilderTxtHex3
			var suf SetUintFs[uint8] = SetUintFsTxtHex3
			var chk NameChecker = NameCheckerNoCheck

			const hname string = "head.txt"
			const tname string = "tail.txt"

			var mbf ManagerBuilderFactoryFs[uint8] = ManagerBuilderFactoryFs[uint8]{}.
				WithGet(guf).
				WithSet(suf).
				WithCheck(chk)

			hmbf, e := mbf.WithName(hname).Build()
			mustNil(e)
			tmbf, e := mbf.WithName(tname).Build()
			mustNil(e)

			var hmu ManagerUint[uint8] = hmbf.BuildManager()
			var tmu ManagerUint[uint8] = tmbf.BuildManager()

			var rmu RingManagerUint[uint8] = RingManagerUintNew(hmu, tmu, root)

			var wb WriteBuilder = WriteBuilder{}.Default()
			wtr, e := wb.BuildNoRename()
			mustNil(e)

			var u2h uint2hex[uint8] = uint2hex3

			wrhbu, e := WriteRequestHandlerBuilderUintNew(rmu, wtr, u2h)
			mustNil(e)
			var wrh WriteRequestHandler = wrhbu.NewHandler()

			var h2u hex2uint[uint8] = hex2uint3

			wehbu, e := WroteEventHandlerBuilderUintNew(h2u, rmu)
			mustNil(e)
			var weh WroteEventHandler = wehbu.NewHandler()

			var htc HeadTailCounter[uint8] = HeadTailCounter3

			var lf ListUint[uint8] = ListUintFallbackNew3
			var lu ListUint[uint8] = rmu.NewList(htc).Fallback(lf)

			var lh ListRequestHandler[uint8] = lu.NewHandler()

			var nbu NameBuilderUint[uint8] = NameBuilderUint3(root)

			var du DeleteUint[uint8] = DeleteUintBuilder(nbu)(chk)
			var dh DeleteHandler[uint8] = du.
				NoentIgnored().
				NewHandler()

			var rbu ReadByUint[uint8] = ReadByUintBuilder(nbu)(chk)
			var rh ReadHandler[uint8] = rbu.
				NoentIgnored().
				NewHandler()

			var rsf RingServiceFactory[uint8] = RingServiceFactory[uint8]{}.
				WithWriteHandler(wrh).
				WithWroteHandler(weh).
				WithListHandler(lh).
				WithDeleteHandler(dh).
				WithReadHandler(rh)

			rs, e := rsf.Build()
			mustNil(e)

			t.Run("service got", func(svc RingService[uint8])func(*testing.T){
				return func(t *testing.T){
					t.Parallel()

					var uw Uint2Writer[uint8] = Uint2WriterHexTxt3.
						WithSuffix([]byte("\n"))
					var lewt ListEventWriterTo[uint8] = uw.NewEventWriter()

					const udel uint8 = 0x37

					t.Run("DeleteRequest", func(t *testing.T){
						t.Parallel()

						var dreq DeleteRequest[uint8] = DeleteRequestNew(udel)
						var evt ServiceEvent = svc.Handle(dreq, lewt)

						t.Run("Must be empty body", check(len(evt.Body()), 0))
						t.Run("Must be ok", check(evt.Status(), StatusOk))
						t.Run("Must not fail", check(nil==evt.Err(), true))
					})
				}
			}(rs))
		})
	})
}
