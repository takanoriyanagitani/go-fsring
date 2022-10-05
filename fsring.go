package fsring

import (
	"errors"
	"fmt"
)

type ServiceStatus int

const (
	// Unexpected status
	StatusUnknown  ServiceStatus = iota

	// No data for specified id
	StatusNotFound               = iota

	StatusOk                     = iota

	// Out of space
	StatusTooMany                = iota

	StatusNg                     = iota
)

type RingService[T any] struct {
	wreqh WriteRequestHandler
	wevth WroteEventHandler
	lrht  ListRequestHandler[T]
	drh   DeleteHandler[T]
	rh    ReadHandler[T]
	reh   RemovedEventHandler[T]
}

type ServiceEvent struct {
	body   []byte
	stat   error
	status ServiceStatus
}

func (s ServiceEvent) Body() []byte          { return s.body }
func (s ServiceEvent) Err() error            { return s.stat }
func (s ServiceEvent) Status() ServiceStatus { return s.status }

func (s ServiceEvent) WithBody(b []byte) ServiceEvent {
	s.body = b
	return s
}

func (s ServiceEvent) WithStat(stat error) ServiceEvent {
	s.stat = stat
	return s
}

func (s ServiceEvent) WithStatus(status ServiceStatus) ServiceEvent {
	s.status = status
	return s
}

func ServiceEventNew(body []byte, stat error) ServiceEvent {
	return ServiceEvent{}.
		WithBody(body).
		WithStat(stat).
		WithStatus(OptUnwrapOrElse(
			func() (status ServiceStatus, hasValue bool) {
				return OptFromBool(nil == stat, func() ServiceStatus { return StatusOk })
			},
			func() ServiceStatus { return StatusNg },
		)())
}

func ServiceEventNg(e error) ServiceEvent {
	return ServiceEvent{}.
		WithStat(e).
		WithStatus(StatusNg)
}

func ServiceEventOk(body []byte) ServiceEvent {
	return ServiceEvent{}.
		WithBody(body).
		WithStatus(StatusOk)
}

type RingServiceFactory[T any] struct {
	WriteRequestHandler
	WroteEventHandler
	ListRequestHandler[T]
	DeleteHandler[T]
	ReadHandler[T]
	RemovedEventHandler[T]
}

func RingServiceFactoryMemDefault3(wtr Write) func(dir string) RingServiceFactory[uint8] {
	return func(dir string) RingServiceFactory[uint8] {
		var hmu ManagerUint[uint8] = ManagerUintMemNew[uint8](0)
		var tmu ManagerUint[uint8] = ManagerUintMemNew[uint8](255)
		var rmu RingManagerUint[uint8] = RingManagerUintNew(hmu, tmu, dir)

		wrhbu, e := WriteRequestHandlerBuilderUintNew(rmu, wtr, uint2hex3)
		var wrh WriteRequestHandler = OptUnwrapOrDefault(
			func() (WriteRequestHandler, bool) {
				return OptFromBool(
					nil == e,
					wrhbu.NewHandler,
				)
			},
		)()

		wehbu, e := WroteEventHandlerBuilderUintNew3(rmu)
		var weh WroteEventHandler = OptUnwrapOrDefault(
			func() (WroteEventHandler, bool) {
				return OptFromBool(
					nil == e,
					wehbu.NewHandler,
				)
			},
		)()

		var lh ListRequestHandler[uint8] = ListUintFallbackNew3.NewHandler()

		var nbu NameBuilderUint[uint8] = NameBuilderUint3(dir)
		var chk NameChecker = NameCheckerNoCheck

		var dh DeleteHandler[uint8] = DeleteUintBuilder(nbu)(chk).NoentIgnored().NewHandler()

		var rh ReadHandler[uint8] = ReadByUintBuilder(nbu)(chk).NoentIgnored().NewHandler()

		rehbu, e := RemovedEventHandlerBuilderUint3New(rmu)
		var reh RemovedEventHandler[uint8] = OptUnwrapOrDefault(
			func() (RemovedEventHandler[uint8], bool) {
				return OptFromBool(
					nil == e,
					rehbu.NewHandler,
				)
			},
		)()

		return RingServiceFactory[uint8]{}.
			WithWriteHandler(wrh).
			WithWroteHandler(weh).
			WithListHandler(lh).
			WithDeleteHandler(dh).
			WithReadHandler(rh).
			WithRemovedHandler(reh)
	}
}

func (f RingServiceFactory[T]) WithWriteHandler(h WriteRequestHandler) RingServiceFactory[T] {
	f.WriteRequestHandler = h
	return f
}

func (f RingServiceFactory[T]) WithWroteHandler(h WroteEventHandler) RingServiceFactory[T] {
	f.WroteEventHandler = h
	return f
}

func (f RingServiceFactory[T]) WithListHandler(h ListRequestHandler[T]) RingServiceFactory[T] {
	f.ListRequestHandler = h
	return f
}

func (f RingServiceFactory[T]) WithDeleteHandler(h DeleteHandler[T]) RingServiceFactory[T] {
	f.DeleteHandler = h
	return f
}

func (f RingServiceFactory[T]) WithReadHandler(h ReadHandler[T]) RingServiceFactory[T] {
	f.ReadHandler = h
	return f
}

func (f RingServiceFactory[T]) WithRemovedHandler(h RemovedEventHandler[T]) RingServiceFactory[T] {
	f.RemovedEventHandler = h
	return f
}

func (f RingServiceFactory[T]) Build() (RingService[T], error) {
	var valid bool = IterFromArr([]bool{
		nil != f.WriteRequestHandler,
		nil != f.WroteEventHandler,
		nil != f.ListRequestHandler,
		nil != f.DeleteHandler,
		nil != f.ReadHandler,
		nil != f.RemovedEventHandler,
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
				reh:   f.RemovedEventHandler,
			}
		},
		func() error { return fmt.Errorf("Invalid factory") },
	)
}

func (s RingService[T]) List(req ListRequest, wtr ListEventWriterTo[T]) ServiceEvent {
	l, e := s.lrht(req)
	return l.ToServiceEvent(e, wtr)
}

func (s RingService[T]) Write(req WriteRequest) ServiceEvent {
	f := ComposeErr(
		s.wreqh,
		func(evt WroteEvent) (WroteEvent, error) { return evt, s.wevth(evt) },
	)
	e := ErrOnly(f)(req)
	return OptUnwrapOrElse(
		func() (evt ServiceEvent, hasValue bool) {
			return OptFromBool(
				errors.Is(e, ErrNonEmpty),
				func() ServiceEvent { return ServiceEventNg(e).WithStatus(StatusTooMany) },
			)
		},
		func() ServiceEvent { return ServiceEventNew(nil, e) },
	)()
}

func (s RingService[T]) Del(req DeleteRequest[T]) ServiceEvent {
	evt, e := ComposeErr(
		s.drh, // DeleteRequest[T] -> RemovedEvent[T], error
		func(evt RemovedEvent[T]) (RemovedEvent[T], error) { return evt, s.reh(evt) },
	)(req)
	return evt.ToServiceEvent(e)
}

func (s RingService[T]) Get(req ReadRequest[T]) ServiceEvent {
	r, e := s.rh(req)
	return r.ToServiceEvent(e)
}

func (s RingService[T]) Handle(req interface{}, wtr ListEventWriterTo[T]) ServiceEvent {
	switch q := req.(type) {
	case ListRequest:
		return s.List(q, wtr)
	case WriteRequest:
		return s.Write(q)
	case DeleteRequest[T]:
		return s.Del(q)
	case ReadRequest[T]:
		return s.Get(q)
	default:
		return ServiceEventNg(fmt.Errorf("Invalid request"))
	}
}
