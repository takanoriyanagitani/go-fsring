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

func (g GetUint[T]) OrElse(f func(error) (T, error)) (T, error) {
	t, e := g()
	if nil != e {
		return f(e)
	}
	return t, nil
}

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
				f, e := os.Create(chk(filename))
				if nil != e {
					return e
				}
				defer func() {
					_ = f.Close() // ignore close error after sync
				}()
				var su SetUint[T] = w2s(f)
				return Err1st([]func() error{
					func() error { return su(neo) },
					func() error { return f.Sync() },
				})
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

func (m ManagerUint[T]) Get() (T, error) { return m.get() }

type RingMangerUint[T uint8 | uint16] struct {
	head ManagerUint[T]
	tail ManagerUint[T]
	dir  string
}

func RingMangerUintNew[T uint8 | uint16](
	head ManagerUint[T],
	tail ManagerUint[T],
	dir string,
) RingMangerUint[T] {
	return RingMangerUint[T]{
		head,
		tail,
		dir,
	}
}

func (r RingMangerUint[T]) next() (T, error) {
	return ComposeErr(
		func(m ManagerUint[T]) (T, error) { return m.get() },
		func(t T) (T, error) { return t + 1, nil },
	)(r.tail)
}

func (r RingMangerUint[T]) nextName(u2h uint2hex[T]) (string, error) {
	return ComposeErr(
		IgnoreArg[RingMangerUint[T]](r.next), // () => T, error
		ErrFuncGen(u2h),                      // T -> string, error
	)(r)
}

func (r RingMangerUint[T]) nextPath(u2h uint2hex[T]) (string, error) {
	name, e := r.nextName(u2h)
	return filepath.Join(r.dir, name), e
}

func (r RingMangerUint[T]) updateTail(neo T) error { return r.tail.set(neo) }
func (r RingMangerUint[T]) UpdateTail(h2u hex2uint[T], neo string) error {
	var f func(string) (T, error) = ComposeErr(
		h2u, // string -> T, error
		func(t T) (T, error) { return t, r.updateTail(t) },
	)
	return ErrOnly(f)(neo)
}

func (r RingMangerUint[T]) Head() (T, error) {
	var h ManagerUint[T] = r.head
	return h.Get()
}

func (r RingMangerUint[T]) Tail() (T, error) {
	var t ManagerUint[T] = r.tail
	return t.Get()
}

func (r RingMangerUint[T]) Count(counter HeadTailCounter[T]) (T, error) {
	h2ht := func(head T) (ht [2]T, e error) {
		tail, e := r.Tail()
		ht[0] = head
		ht[1] = tail
		return
	}
	r2h := func(m RingMangerUint[T]) (T, error) { return m.Head() }
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
