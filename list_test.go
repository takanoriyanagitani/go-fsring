package fsring

import (
	"os"
	"path/filepath"
	"testing"
)

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("count2listUint", func(t *testing.T) {
		t.Parallel()

		t.Run("zero", func(t *testing.T) {
			t.Parallel()

			var i Iter[uint8] = count2listUint[uint8](0, 42)
			_, hasValue := i()
			t.Run("Must be empty", check(hasValue, false))
		})

		t.Run("single item", func(t *testing.T) {
			t.Parallel()

			var i Iter[uint8] = count2listUint[uint8](1, 42)
			u, hasValue := i()
			t.Run("Must not be empty", check(hasValue, true))
			t.Run("Must be same", check(u, 42))
		})

		t.Run("many integers", func(t *testing.T) {
			t.Parallel()

			var i Iter[uint8] = count2listUint[uint8](3, 42)
			var tot uint8 = i.Reduce(0, func(a, b uint8) uint8 { return a + b })
			t.Run("Must be same", check(tot, 42+43+44))
		})

		t.Run("wraparound", func(t *testing.T) {
			t.Parallel()

			var i Iter[uint8] = count2listUint[uint8](3, 254)
			var tot uint8 = i.Reduce(0, func(a, b uint8) uint8 { return a + b })
			t.Run("Must be same", check(tot, 253)) // 254+255+0 = 254-1+256 = 253
		})
	})

	t.Run("sub3", func(t *testing.T) {
		t.Parallel()

		t.Run("zero", func(t *testing.T) {
			t.Parallel()

			d, b := sub3(0, 0, 0)
			t.Run("must be same(diff)", check(d, 0))
			t.Run("must be same(borrow)", check(b, 0))
		})

		t.Run("no borrow", func(t *testing.T) {
			t.Parallel()

			d, b := sub3(43, 1, 0)
			t.Run("must be same(diff)", check(d, 42))
			t.Run("must be same(borrow)", check(b, 0))
		})

		t.Run("borrow", func(t *testing.T) {
			t.Parallel()

			d, b := sub3(0, 1, 0)
			t.Run("must be same(diff)", check(d, 255))
			t.Run("must be same(borrow)", check(b, 1))
		})
	})

	t.Run("HeadTailCounter3", func(t *testing.T) {
		t.Parallel()

		t.Run("minimum", func(t *testing.T) {
			t.Parallel()

			var cnt uint8 = HeadTailCounter3(0, 0)
			t.Run("Must be same", check(cnt, 0))
		})

		t.Run("no borrow", func(t *testing.T) {
			t.Parallel()

			var cnt uint8 = HeadTailCounter3(43, 1)
			t.Run("Must be same", check(cnt, 42))
		})

		t.Run("borrow", func(t *testing.T) {
			t.Parallel()

			var cnt uint8 = HeadTailCounter3(0, 1)
			t.Run("Must be same", check(cnt, 255))
		})
	})

	t.Run("ListUint", func(t *testing.T) {
		t.Parallel()

		var ITEST_FSRING_DIR string = os.Getenv("ITEST_FSRING_DIR")
		if len(ITEST_FSRING_DIR) < 1 {
			t.Skip("skipping tests using filesystem...")
		}

		t.Run("RingManagerUint", func(t *testing.T) {
			t.Parallel()
			var wb WriteBuilder = WriteBuilder{}.Default()
			w, e := wb.BuildNoRename()
			mustNil(e)

			t.Run("write got", func(wtr Write) func(*testing.T) {
				return func(t *testing.T) {
					t.Parallel()

					t.Run("Uint8", func(t *testing.T) {
						t.Parallel()

						var root string = filepath.Join(ITEST_FSRING_DIR, "List/Uint8")
						e := os.MkdirAll(root, 0755)
						mustNil(e)

						var head string = "head.txt"
						var tail string = "tail.txt"

						var mbf ManagerBuilderFactoryFs[uint8] = ManagerBuilderFactoryFs[uint8]{}.
							WithCheck(NameCheckerNoCheck).
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

						var rm RingManagerUint[uint8] = RingManagerUintNew(mh, mt, root)

						t.Run("manager got", func(mng RingManagerUint[uint8]) func(*testing.T) {
							return func(t *testing.T) {
								t.Parallel()

								var counter HeadTailCounter[uint8] = HeadTailCounter3
								var l ListUint[uint8] = mng.NewList(counter)

								var h2u hex2uint[uint8] = hex2uint3

								t.Run("minimum", func(t *testing.T) {
									e := mng.UpdateHead(h2u, "00")
									mustNil(e)

									e = mng.UpdateTail(h2u, "00")
									mustNil(e)

									names, e := l()
									mustNil(e)

									var cnt int = names.Count()
									t.Run("Must same(count)", check(cnt, 1))

									names, e = l()
									mustNil(e)
									add := func(a, b uint8) uint8 { return a + b }
									var tot uint8 = names.Reduce(0, add)
									t.Run("Must same(tot)", check(tot, 0))
								})

								t.Run("no borrow", func(t *testing.T) {
									e := mng.UpdateHead(h2u, "00")
									mustNil(e)

									e = mng.UpdateTail(h2u, "29")
									mustNil(e)

									names, e := l()
									mustNil(e)

									var cnt int = names.Count()
									t.Run("Must same(count)", check(cnt, 42))

									names, e = l()
									mustNil(e)

									var na []uint8 = names.ToArray()
									t.Run("len check", check(len(na), 42))

									t.Run("1st item", check(na[0], 0))
									t.Run("2nd item", check(na[1], 1))
									t.Run("3rd item", check(na[2], 2))
									t.Run("last item", check(na[41], 41))
								})

								t.Run("borrow", func(t *testing.T) {
									e := mng.UpdateHead(h2u, "40")
									mustNil(e)

									e = mng.UpdateTail(h2u, "3f")
									mustNil(e)

									names, e := l()
									mustNil(e)

									var cnt int = names.Count()
									t.Run("Must same(count)", check(cnt, 256))

									names, e = l()
									mustNil(e)

									var na []uint8 = names.ToArray()
									t.Run("len check", check(len(na), 256))

									t.Run("1st item", check(na[0], 0x40))
									t.Run("2nd item", check(na[1], 0x41))
									t.Run("3rd item", check(na[2], 0x42))
									t.Run("last item", check(na[255], 0x3f))
								})
							}
						}(rm))
					})
				}
			}(w))
		})
	})

	t.Run("ListUintFallback", func(t *testing.T) {
		t.Parallel()

		t.Run("uint8", func(t *testing.T) {
			t.Parallel()

			var lu ListUint[uint8] = ListUintFallbackNew3
			names, e := lu()
			mustNil(e)
			var cnt int = names.Count()
			t.Run("Must be same", check(cnt, 256))
		})
	})
}
