package fsring

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type GetUint[T uint8 | uint16] func() (T, error)
type SetUint[T uint8 | uint16] func(neo T) error

func (g GetUint[T]) IgnoreNoent(f func() (T, error)) (T, error) {
	t, e := g()
	if nil != e {
		ef := ErrIgnore(func(err error) bool { return errors.Is(err, fs.ErrNotExist) })
		e = ef(e)
		if nil != e {
			return 0, e
		}
		return f()
	}
	return t, nil
}

func (g GetUint[T]) NoentIgnored(f func() (T, error)) GetUint[T] {
	return func() (T, error) {
		return g.IgnoreNoent(f)
	}
}

func (g GetUint[T]) Fallback(alt T) GetUint[T] {
	return func() (T, error) {
		t, e := g()
		if nil != e {
			return alt, nil
		}
		return t, nil
	}
}

type uint2hex[T uint8 | uint16] func(T) string

var uint2hex3 uint2hex[uint8] = uint2hexBuilder[uint8]("%02x")
var uint2hex4 uint2hex[uint16] = uint2hexBuilder[uint16]("%04x")

func getUintNewTxt[T uint8 | uint16](conv func(string) (T, error)) func(string) GetUint[T] {
	return func(s string) GetUint[T] {
		return func() (T, error) {
			return ComposeErr(
				conv,
				ErrFuncGen(Identity[T]),
			)(s)
		}
	}
}

var getUintNewTxtHex3 func(string) GetUint[uint8] = getUintNewTxt(hex2uint3)
var getUintNewTxtHex4 func(string) GetUint[uint16] = getUintNewTxt(hex2uint4)

type GetUintFs[T uint8 | uint16] func(NameChecker) func(filename string) GetUint[T]

func GetUintFsBuilderTxt[T uint8 | uint16](s2g func(string) GetUint[T]) GetUintFs[T] {
	return func(chk NameChecker) func(string) GetUint[T] {
		return func(filename string) GetUint[T] {
			return func() (T, error) {
				return ComposeErr(
					os.ReadFile, // string -> []byte, error
					func(b []byte) (T, error) { return s2g(string(b))() },
				)(chk(filename))
			}
		}
	}
}

var GetUintFsBuilderTxtHex3 GetUintFs[uint8] = GetUintFsBuilderTxt(getUintNewTxtHex3)
var GetUintFsBuilderTxtHex4 GetUintFs[uint16] = GetUintFsBuilderTxt(getUintNewTxtHex4)

func uint2hexBuilder[T uint8 | uint16](format string) func(T) string {
	return func(t T) string { return fmt.Sprintf(format, t) }
}

func setUintWriteNew[T uint8 | uint16](conv func(T) string) func(io.Writer) SetUint[T] {
	return func(w io.Writer) SetUint[T] {
		f := ComposeErr(
			ErrFuncGen(conv), // T -> string
			func(s string) (int, error) { return w.Write([]byte(s)) },
		)
		return ErrOnly(f)
	}
}

var setUintWriterTxtHex3 func(io.Writer) SetUint[uint8] = setUintWriteNew(uint2hex3)
var setUintWriterTxtHex4 func(io.Writer) SetUint[uint16] = setUintWriteNew(uint2hex4)

type SetUintFs[T uint8 | uint16] func(NameChecker) func(filename string) SetUint[T]

func SetUintFsBuilder[T uint8 | uint16](w2s func(io.Writer) SetUint[T]) SetUintFs[T] {
	return func(chk NameChecker) func(string) SetUint[T] {
		return func(filename string) SetUint[T] {
			return func(neo T) error {
				f, e := ComposeErr(
					func(d string) (string, error) { return filename, os.MkdirAll(chk(d), 0750) },
					func(name string) (*os.File, error) { return os.Create(chk(name)) },
				)(filepath.Dir(filename))
				if nil != e {
					return e
				}
				defer func() {
					_ = f.Close() // ignore close error(allow head/tail to be dirty)
				}()
				var su SetUint[T] = w2s(f)
				return su(neo)
			}
		}
	}
}

var SetUintFsTxtHex3 SetUintFs[uint8] = SetUintFsBuilder(setUintWriterTxtHex3)
var SetUintFsTxtHex4 SetUintFs[uint16] = SetUintFsBuilder(setUintWriterTxtHex4)

type ManagerBuilderFs[T uint8 | uint16] struct {
	get GetUintFs[T]
	set SetUintFs[T]
	chk NameChecker
	nam string
}

func (b ManagerBuilderFs[T]) Build() (GetUint[T], SetUint[T]) {
	g := b.get(b.chk)(b.nam)
	s := b.set(b.chk)(b.nam)
	return g, s
}

func (b ManagerBuilderFs[T]) BuildManager() ManagerUint[T] {
	get, set := b.Build()
	return ManagerUint[T]{
		get,
		set,
	}
}

type ManagerBuilderFactoryFs[T uint8 | uint16] struct {
	GetUintFs[T]
	SetUintFs[T]
	NameChecker
	Filename string
}

func (f ManagerBuilderFactoryFs[T]) WithGet(g GetUintFs[T]) ManagerBuilderFactoryFs[T] {
	f.GetUintFs = g
	return f
}

func (f ManagerBuilderFactoryFs[T]) WithSet(s SetUintFs[T]) ManagerBuilderFactoryFs[T] {
	f.SetUintFs = s
	return f
}

func (f ManagerBuilderFactoryFs[T]) WithCheck(chk NameChecker) ManagerBuilderFactoryFs[T] {
	f.NameChecker = chk
	return f
}

func (f ManagerBuilderFactoryFs[T]) WithName(name string) ManagerBuilderFactoryFs[T] {
	f.Filename = name
	return f
}

func (f ManagerBuilderFactoryFs[T]) Build() (b ManagerBuilderFs[T], e error) {
	var valid bool = IterFromArr([]bool{
		nil != f.GetUintFs,
		nil != f.SetUintFs,
		nil != f.NameChecker,
		0 < len(f.Filename),
	}).All(Identity[bool])
	return ErrFromBool(
		valid,
		func() ManagerBuilderFs[T] {
			return ManagerBuilderFs[T]{
				get: f.GetUintFs,
				set: f.SetUintFs,
				chk: f.NameChecker,
				nam: f.Filename,
			}
		},
		func() error { return fmt.Errorf("Invalid factory") },
	)
}

type ManagerUint[T uint8 | uint16] struct {
	get GetUint[T]
	set SetUint[T]
}

func (m ManagerUint[T]) NoentIgnored(f func() (T, error)) ManagerUint[T] {
	m.get = m.get.NoentIgnored(f)
	return m
}

func (m ManagerUint[T]) Fallback(alt T) ManagerUint[T] {
	m.get = m.get.Fallback(alt)
	return m
}

func (m ManagerUint[T]) Get() (T, error) { return m.get() }

type RingManagerUint[T uint8 | uint16] struct {
	head ManagerUint[T]
	tail ManagerUint[T]
	dir  string
}

func RingManagerUintNew[T uint8 | uint16](
	head ManagerUint[T],
	tail ManagerUint[T],
	dir string,
) RingManagerUint[T] {
	return RingManagerUint[T]{
		head,
		tail,
		dir,
	}
}

func (r RingManagerUint[T]) next() (T, error) {
	return ComposeErr(
		func(m ManagerUint[T]) (T, error) { return m.get() },
		func(t T) (T, error) { return t + 1, nil },
	)(r.tail)
}

func (r RingManagerUint[T]) nextName(u2h uint2hex[T]) (string, error) {
	return ComposeErr(
		IgnoreArg[RingManagerUint[T]](r.next), // () => T, error
		ErrFuncGen(u2h),                       // T -> string, error
	)(r)
}

func (r RingManagerUint[T]) nextPath(u2h uint2hex[T]) (string, error) {
	name, e := r.nextName(u2h)
	return filepath.Join(r.dir, name), e
}

func (r RingManagerUint[T]) updateHead(neo T) error { return r.head.set(neo) }
func (r RingManagerUint[T]) updateTail(neo T) error { return r.tail.set(neo) }

func (r RingManagerUint[T]) Refresh(head, tail ManagerUint[T]) RingManagerUint[T] {
	r.head = head
	r.tail = tail
	return r
}

func (r RingManagerUint[T]) UpdateTail(h2u hex2uint[T], neo string) error {
	var f func(string) (T, error) = ComposeErr(
		h2u, // string -> T, error
		func(t T) (T, error) { return t, r.updateTail(t) },
	)
	return ErrOnly(f)(neo)
}

func (r RingManagerUint[T]) UpdateHead(h2u hex2uint[T], neo string) error {
	var f func(string) (T, error) = ComposeErr(
		h2u, // string -> T, error
		func(t T) (T, error) { return t, r.updateHead(t) },
	)
	return ErrOnly(f)(neo)
}

func (r RingManagerUint[T]) Head() (T, error) {
	var h ManagerUint[T] = r.head
	return h.Get()
}

func (r RingManagerUint[T]) Tail() (T, error) {
	var t ManagerUint[T] = r.tail
	return t.Get()
}

func (r RingManagerUint[T]) count(counter HeadTailCounter[T]) (T, error) {
	h2ht := func(head T) (ht [2]T, e error) {
		tail, e := r.Tail()
		ht[0] = head
		ht[1] = tail
		return
	}
	r2h := func(m RingManagerUint[T]) (T, error) { return m.Head() }
	r2ht := ComposeErr(
		r2h,
		h2ht,
	)
	ht2diff := func(ht [2]T) (diff T) {
		var head T = ht[0]
		var tail T = ht[1]
		return counter(tail, head)
	}
	r2diff := ComposeErr(
		r2ht,
		ErrFuncGen(ht2diff),
	)
	return r2diff(r)
}

func (r RingManagerUint[T]) count2names(cnt int) (names Iter[T], e error) {
	var lbi2names func(lbi T) Iter[T] = Curry(count2listUint[T])(cnt)
	return ComposeErr(
		func(_ RingManagerUint[T]) (T, error) { return r.Head() },
		ErrFuncGen(lbi2names),
	)(r)
}

func (r RingManagerUint[T]) Names(counter HeadTailCounter[T]) (names Iter[T], e error) {
	var counter2count func(HeadTailCounter[T]) (int, error) = ComposeErr(
		r.count,
		func(cnt T) (int, error) { return int(cnt) + 1, nil },
	)
	return ComposeErr(
		counter2count,
		r.count2names,
	)(counter)
}

func (r RingManagerUint[T]) NewList(counter HeadTailCounter[T]) ListUint[T] {
	return func() (names Iter[T], e error) {
		return r.Names(counter)
	}
}
