package fsring

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
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

var checkBytes func(got, expected []byte) func(*testing.T) = checkBuilder(
	func(a, b []byte) (same bool) { return 0 == bytes.Compare(a, b) },
)

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

		t.Run("valid factory", func(t *testing.T) {
			t.Parallel()

			var ITEST_FSRING_DIR string = os.Getenv("ITEST_FSRING_DIR")
			if len(ITEST_FSRING_DIR) < 1 {
				t.Skip("skipping tests using file system")
			}

			var root string = filepath.Join(ITEST_FSRING_DIR, "fsring/valid/factory")

			e := os.RemoveAll(root)
			mustNil(e)
			e = os.MkdirAll(root, 0755)
			mustNil(e)

			var guf GetUintFs[uint8] = GetUintFsBuilderTxtHex3
			var suf SetUintFs[uint8] = SetUintFsTxtHex3
			var chk NameChecker = NameCheckerNoCheck

			var hname string = filepath.Join(root, "head.txt")
			var tname string = filepath.Join(root, "tail.txt")

			e = os.WriteFile(hname, nil, 0644)
			mustNil(e)
			e = os.WriteFile(tname, nil, 0644)
			mustNil(e)

			var mbf ManagerBuilderFactoryFs[uint8] = ManagerBuilderFactoryFs[uint8]{}.
				WithGet(guf).
				WithSet(suf).
				WithCheck(chk)

			hmbf, e := mbf.WithName(hname).Build()
			mustNil(e)
			tmbf, e := mbf.WithName(tname).Build()
			mustNil(e)

			noent0 := func() (uint8, error) { return 0, nil }

			var hmu ManagerUint[uint8] = hmbf.BuildManager().NoentIgnored(noent0).Fallback(0)
			var tmu ManagerUint[uint8] = tmbf.BuildManager().NoentIgnored(noent0).Fallback(255)

			var rmu RingManagerUint[uint8] = RingManagerUintNew(hmu, tmu, root)

			var wb WriteBuilder = WriteBuilder{}.
				Default().
				WithFileSync(FileSyncData)
			wtr, e := wb.BuildNoRename()
			mustNil(e)

			var ie IsEmpty = IsEmptyBuilderNew(chk)
			var w2empty Write = wtr.RejectNonEmpty(ie)

			var u2h uint2hex[uint8] = uint2hex3

			wrhbu, e := WriteRequestHandlerBuilderUintNew(rmu, w2empty, u2h)
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

			rehbu, e := RemovedEventHandlerBuilderUint3New(rmu)
			mustNil(e)

			var reh RemovedEventHandler[uint8] = rehbu.NewHandler()

			var rsf RingServiceFactory[uint8] = RingServiceFactory[uint8]{}.
				WithWriteHandler(wrh).
				WithWroteHandler(weh).
				WithListHandler(lh).
				WithDeleteHandler(dh).
				WithReadHandler(rh).
				WithRemovedHandler(reh)

			rs, e := rsf.Build()
			mustNil(e)

			t.Run("service got", func(svc RingService[uint8]) func(*testing.T) {
				return func(t *testing.T) {
					t.Parallel()

					var uw Uint2Writer[uint8] = Uint2WriterHexTxt3.
						WithSuffix([]byte("\n"))
					var lewt ListEventWriterTo[uint8] = uw.NewEventWriter()

					const udel uint8 = 0x37
					const s404 uint8 = 0x44

					t.Run("DeleteRequest", func(t *testing.T) {
						t.Parallel()

						var dreq DeleteRequest[uint8] = DeleteRequestNew(udel)
						var evt ServiceEvent = svc.Handle(dreq, lewt)

						t.Run("Must be empty body", check(len(evt.Body()), 0))
						t.Run("Must be ok", check(evt.Status(), StatusOk))
						t.Run("Must not fail", check(nil == evt.Err(), true))
					})

					t.Run("ReadRequest", func(t *testing.T) {
						t.Parallel()

						var req ReadRequest[uint8] = ReadRequestNew(s404)
						var evt ServiceEvent = svc.Handle(req, lewt)

						t.Run("Must be empty", check(len(evt.Body()), 0))
						t.Run("Must be noent", check(evt.Status(), StatusNotFound))
						t.Run("Must not fail", check(nil == evt.Err(), true))
					})

					t.Run("Invalid request", func(t *testing.T) {
						t.Parallel()

						var evt ServiceEvent = svc.Handle(nil, lewt)

						t.Run("Must fail", check(nil != evt.Err(), true))
					})

					t.Run("Story", func(t *testing.T) {
						t.Parallel()

						var wreq WriteRequest = WriteRequestNew([]byte("hw"))
						var wevt ServiceEvent = svc.Handle(wreq, lewt)
						t.Run("Must be empty", check(len(wevt.Body()), 0))
						t.Run("Must be ok(write)", check(wevt.Status(), StatusOk))
						t.Run("Must not fail(write)", check(nil == wevt.Err(), true))

						var lreq ListRequest = ListRequest{}
						var levt ServiceEvent = svc.Handle(lreq, lewt)
						t.Run("Must be ok(list)", check(levt.Status(), StatusOk))
						t.Run("Must not fail(list)", check(nil == levt.Err(), true))

						var lf []byte = []byte("\n")
						var keys Iter[[]byte] = IterFromArr(bytes.Split(levt.Body(), lf)).
							Filter(func(item []byte) bool { return 0 < len(item) })
						var mapd Iter[uint8] = IterMap(keys, func(b []byte) uint8 {
							var s string = string(b)
							u, e := hex2uint3(s)
							mustNil(e)
							return u
						})

						var requests Iter[ReadRequest[uint8]] = IterMap(mapd, ReadRequestNew[uint8])
						var events []ServiceEvent = IterMap(
							requests,
							func(q ReadRequest[uint8]) ServiceEvent { return svc.Handle(q, lewt) },
						).Filter(func(evt ServiceEvent) bool {
							return StatusNotFound != evt.Status()
						}).ToArray()
						t.Run("Single event", check(len(events), 1))

						var evt ServiceEvent = events[0]
						t.Run("Must not fail(read)", check(nil == evt.Err(), true))
						t.Run("Must be ok(read)", check(evt.Status(), StatusOk))
						t.Run("Must be same(read)", checkBytes(evt.Body(), []byte("hw")))

						var wrequests Iter[WriteRequest] = IterMap(
							IterInts(0, 1024), func(_ int) WriteRequest {
								return WriteRequestNew(nil)
							},
						)
						var wevents []ServiceEvent = IterMap(
							wrequests, func(q WriteRequest) ServiceEvent {
								return svc.Handle(q, lewt)
							},
						).
							Filter(func(evt ServiceEvent) bool { return nil != evt.Err() }).
							ToArray()
						t.Run("Must fail(too many)", check(0 < len(wevents), true))
						var wfail ServiceEvent = wevents[0]
						t.Run("Must be empty(too many)", check(len(wfail.Body()), 0))
						t.Run("Must be same(too many)", check(wfail.Status(), StatusTooMany))
					})
				}
			}(rs))
		})

		t.Run("mem manager", func(t *testing.T) {
			t.Parallel()

			var ITEST_FSRING_DIR string = os.Getenv("ITEST_FSRING_DIR")
			if len(ITEST_FSRING_DIR) < 1 {
				t.Skip("skipping tests using file system")
			}

			var root string = filepath.Join(ITEST_FSRING_DIR, "fsring/mem/manager")

			e := os.RemoveAll(root)
			mustNil(e)
			e = os.MkdirAll(root, 0755)
			mustNil(e)

			var chk NameChecker = NameCheckerNoCheck

			var hname string = filepath.Join(root, "head.txt")
			var tname string = filepath.Join(root, "tail.txt")

			e = os.WriteFile(hname, nil, 0644)
			mustNil(e)
			e = os.WriteFile(tname, nil, 0644)
			mustNil(e)

			var hmu ManagerUint[uint8] = ManagerUintMemNew[uint8](0)
			var tmu ManagerUint[uint8] = ManagerUintMemNew[uint8](255)

			var rmu RingManagerUint[uint8] = RingManagerUintNew(hmu, tmu, root).
				Refresh(
					ManagerUintMemNew[uint8](0),
					ManagerUintMemNew[uint8](255),
				)

			var wb WriteBuilder = WriteBuilder{}.
				Default().
				WithFileSync(FileSyncData)
			wtr, e := wb.BuildNoRename()
			mustNil(e)

			var ie IsEmpty = IsEmptyBuilderNew(chk)
			var w2empty Write = wtr.RejectNonEmpty(ie)

			var u2h uint2hex[uint8] = uint2hex3

			wrhbu, e := WriteRequestHandlerBuilderUintNew(rmu, w2empty, u2h)
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

			rehbu, e := RemovedEventHandlerBuilderUint3New(rmu)
			mustNil(e)

			var reh RemovedEventHandler[uint8] = rehbu.NewHandler()

			var rsf RingServiceFactory[uint8] = RingServiceFactory[uint8]{}.
				WithWriteHandler(wrh).
				WithWroteHandler(weh).
				WithListHandler(lh).
				WithDeleteHandler(dh).
				WithReadHandler(rh).
				WithRemovedHandler(reh)

			rs, e := rsf.Build()
			mustNil(e)

			t.Run("service got", func(svc RingService[uint8]) func(*testing.T) {
				return func(t *testing.T) {
					t.Parallel()

					var uw Uint2Writer[uint8] = Uint2WriterHexTxt3.
						WithSuffix([]byte("\n"))
					var lewt ListEventWriterTo[uint8] = uw.NewEventWriter()

					const udel uint8 = 0x37
					const s404 uint8 = 0x44

					t.Run("DeleteRequest", func(t *testing.T) {
						t.Parallel()

						var dreq DeleteRequest[uint8] = DeleteRequestNew(udel)
						var evt ServiceEvent = svc.Handle(dreq, lewt)

						t.Run("Must be empty body", check(len(evt.Body()), 0))
						t.Run("Must be ok", check(evt.Status(), StatusOk))
						t.Run("Must not fail", check(nil == evt.Err(), true))
					})

					t.Run("ReadRequest", func(t *testing.T) {
						t.Parallel()

						var req ReadRequest[uint8] = ReadRequestNew(s404)
						var evt ServiceEvent = svc.Handle(req, lewt)

						t.Run("Must be empty", check(len(evt.Body()), 0))
						t.Run("Must be noent", check(evt.Status(), StatusNotFound))
						t.Run("Must not fail", check(nil == evt.Err(), true))
					})

					t.Run("Invalid request", func(t *testing.T) {
						t.Parallel()

						var evt ServiceEvent = svc.Handle(nil, lewt)

						t.Run("Must fail", check(nil != evt.Err(), true))
					})

					t.Run("Story", func(t *testing.T) {
						t.Parallel()

						var wreq WriteRequest = WriteRequestNew([]byte("hw"))
						var wevt ServiceEvent = svc.Handle(wreq, lewt)
						t.Run("Must be empty", check(len(wevt.Body()), 0))
						t.Run("Must be ok(write)", check(wevt.Status(), StatusOk))
						t.Run("Must not fail(write)", check(nil == wevt.Err(), true))

						var lreq ListRequest = ListRequest{}
						var levt ServiceEvent = svc.Handle(lreq, lewt)
						t.Run("Must be ok(list)", check(levt.Status(), StatusOk))
						t.Run("Must not fail(list)", check(nil == levt.Err(), true))

						var lf []byte = []byte("\n")
						var keys Iter[[]byte] = IterFromArr(bytes.Split(levt.Body(), lf)).
							Filter(func(item []byte) bool { return 0 < len(item) })
						var mapd Iter[uint8] = IterMap(keys, func(b []byte) uint8 {
							var s string = string(b)
							u, e := hex2uint3(s)
							mustNil(e)
							return u
						})

						var requests Iter[ReadRequest[uint8]] = IterMap(mapd, ReadRequestNew[uint8])
						var events []ServiceEvent = IterMap(
							requests,
							func(q ReadRequest[uint8]) ServiceEvent { return svc.Handle(q, lewt) },
						).Filter(func(evt ServiceEvent) bool {
							return StatusNotFound != evt.Status()
						}).ToArray()
						t.Run("Single event", check(len(events), 1))

						var evt ServiceEvent = events[0]
						t.Run("Must not fail(read)", check(nil == evt.Err(), true))
						t.Run("Must be ok(read)", check(evt.Status(), StatusOk))
						t.Run("Must be same(read)", checkBytes(evt.Body(), []byte("hw")))

						var wrequests Iter[WriteRequest] = IterMap(
							IterInts(0, 1024), func(_ int) WriteRequest {
								return WriteRequestNew(nil)
							},
						)
						var wevents []ServiceEvent = IterMap(
							wrequests, func(q WriteRequest) ServiceEvent {
								return svc.Handle(q, lewt)
							},
						).
							Filter(func(evt ServiceEvent) bool { return nil != evt.Err() }).
							ToArray()
						t.Run("Must fail(too many)", check(0 < len(wevents), true))
						var wfail ServiceEvent = wevents[0]
						t.Run("Must be empty(too many)", check(len(wfail.Body()), 0))
						t.Run("Must be same(too many)", check(wfail.Status(), StatusTooMany))
					})
				}
			}(rs))
		})
	})
}
