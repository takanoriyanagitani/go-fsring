package fsring

import (
	"bytes"
	"context"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func BenchmarkAll(b *testing.B) {
	b.Run("RingService", func(b *testing.B) {
		var ITEST_FSRING_DIR string = os.Getenv("ITEST_FSRING_DIR")
		if len(ITEST_FSRING_DIR) < 1 {
			b.Skip("skipping benchmark using file system")
		}

		var root string = filepath.Join(ITEST_FSRING_DIR, "Bench/RingService")

		e := os.RemoveAll(root)
		mustNil(e)
		e = os.MkdirAll(root, 0755)
		mustNil(e)

		var wtr Write = WriteNocheckFdatasync
		var rsf RingServiceFactory[uint8] = RingServiceFactoryMemDefault3(wtr)(root)
		rs, e := rsf.Build()
		mustNil(e)

		b.Run("service got", func(svc RingService[uint8]) func(*testing.B) {
			return func(b *testing.B) {
				var uw Uint2Writer[uint8] = Uint2WriterHexTxt3.
					WithSuffix([]byte("\n"))
				var lewt ListEventWriterTo[uint8] = uw.NewEventWriter()

				getBytesBuilder := func(sz int) func() []byte {
					buf := make([]byte, sz)
					return func() []byte {
						_, e := rand.Read(buf)
						mustNil(e)
						return buf
					}
				}

				chk := func(sz int, wait time.Duration) func(*testing.B) {
					var genBytes func() []byte = getBytesBuilder(sz)
					return func(b *testing.B) {
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						var lock sync.Mutex

						getList := func(req ListRequest) ServiceEvent {
							lock.Lock()
							defer lock.Unlock()
							return svc.Handle(req, lewt)
						}

						getBody := func(req ReadRequest[uint8]) ServiceEvent {
							lock.Lock()
							defer lock.Unlock()
							return svc.Handle(req, lewt)
						}

						remove := func(req DeleteRequest[uint8]) ServiceEvent {
							lock.Lock()
							defer lock.Unlock()
							return svc.Handle(req, lewt)
						}

						write := func(req WriteRequest) ServiceEvent {
							lock.Lock()
							defer lock.Unlock()
							return svc.Handle(req, lewt)
						}

						go func() {
							for {
								select {
								case <-ctx.Done():
									return
								default:
									func() {
										var levt ServiceEvent = getList(ListRequest{})
										mustNil(levt.Err())
										var lbody []byte = levt.Body()
										lf := []byte("\n")
										var keys Iter[[]byte] = IterFromArr(bytes.Split(lbody, lf)).
											Filter(func(b []byte) bool { return 0 < len(b) })
										var mapd Iter[uint8] = IterMap(keys, func(ba []byte) uint8 {
											u, e := hex2uint3(string(ba))
											mustNil(e)
											return u
										})
										var reads []ReadRequest[uint8] = IterMap(
											mapd, ReadRequestNew[uint8],
										).ToArray()
										for _, r := range reads {
											var revt ServiceEvent = getBody(r)
											mustNil(revt.Err())
											if StatusOk == revt.Status() {
												dreq := DeleteRequestNew(r.Target())
												var devt ServiceEvent = remove(dreq)
												mustNil(devt.Err())
											}
										}

										time.Sleep(10 * wait)
									}()
								}
							}
						}()

						b.ResetTimer()

						for i := 0; i < b.N; i++ {
							var b []byte = genBytes()
							var req WriteRequest = WriteRequestNew(b)

							evt := write(req)

							mustNil(evt.Err())
							time.Sleep(wait)
						}
					}
				}

				b.Run("8 KiB", chk(8192, 10*time.Millisecond))
				b.Run("80 KiB", chk(81920, 100*time.Millisecond))
				b.Run("800 KiB", chk(819200, 1000*time.Millisecond))

			}
		}(rs))
	})
}
