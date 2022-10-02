package fsring

import (
	"fmt"
)

type RingService[T any] struct {
	wreqh WriteRequestHandler
	wevth WroteEventHandler
	lrht  ListRequestHandler[T]
	drh   DeleteHandler[T]
	rh    ReadHandler[T]
}

type RingServiceFactory[T any] struct {
	WriteRequestHandler
	WroteEventHandler
	ListRequestHandler[T]
	DeleteHandler[T]
	ReadHandler[T]
}

func (f RingServiceFactory[T]) Build() (RingService[T], error) {
	var valid bool = IterFromArr([]bool{
		nil != f.WriteRequestHandler,
		nil != f.WroteEventHandler,
		nil != f.ListRequestHandler,
		nil != f.DeleteHandler,
		nil != f.ReadHandler,
	}).All(Identity[bool])
	return ErrFromBool(
		valid,
		func() RingService[T] {
			return RingService[T]{
				wreqh: f.WriteRequestHandler,
				wevth: f.WroteEventHandler,
				lrht:  f.ListRequestHandler,
				drh:   f.DeleteHandler,
				rh:    f.ReadHandler,
			}
		},
		func() error { return fmt.Errorf("Invalid factory") },
	)
}

func (s RingService[T]) List(req ListRequest) (ListEvent[T], error) { return s.lrht(req) }

func (s RingService[T]) Write(req WriteRequest) error {
	f := ComposeErr(
		s.wreqh,
		func(evt WroteEvent) (WroteEvent, error) { return evt, s.wevth(evt) },
	)
	return ErrOnly(f)(req)
}

func (s RingService[T]) Del(req DeleteRequest[T]) (RemovedEvent, error) { return s.drh(req) }

func (s RingService[T]) Get(req ReadRequest[T]) (ReadEvent, error) { return s.rh(req) }
