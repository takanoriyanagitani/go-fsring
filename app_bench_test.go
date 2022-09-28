package fsring

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"math/rand"
)

func BenchmarkApp(b *testing.B) {
	var ITEST_FSRING_DIR string = os.Getenv("ITEST_FSRING_DIR")

	if len(ITEST_FSRING_DIR) < 1 {
		b.Skip("skipping benchmarks using filesystem")
	}

	b.Run("HandleWriteRequest", func(b *testing.B) {
		var root string = filepath.Join(ITEST_FSRING_DIR, "BenchmarkApp/HandleWriteRequest")
		var mng string = "mng.txt"
		var chk NameChecker = NameCheckerNoCheck

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

		rgen := func(b []byte){
			_, e := rand.Read(b)
			mustNil(e)
		}

		b.Run("app got", func(a App) func(*testing.B){
			return func(b *testing.B){
				b.Run("empty", func(b *testing.B){
					b.ResetTimer()

					for i := 0; i < b.N; i++ {
						e := a.HandleWriteRequest(WriteRequest{ data: nil })
						mustNil(e)
					}
				})

				rbench := func(bsz int) func(*testing.B){
					return func(b *testing.B){
						var buf []byte = make([]byte, bsz)
						b.ResetTimer()
						for i:=0; i<b.N; i++{
							rgen(buf)
							e := a.HandleWriteRequest(WriteRequest{
								data: buf,
							})
							mustNil(e)
						}
					}
				}

				b.Run("bsz=8192bytes", rbench(8192))
				b.Run("bsz=81920bytes", rbench(81920))
			}
		}(app))
	})
}
