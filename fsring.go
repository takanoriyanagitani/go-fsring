package fsring

import (
	"fmt"
)

type ServiceStatus int

const (
	StatusUnknown  ServiceStatus = iota
	StatusNotFound               = iota
	StatusOk                     = iota
	StatusTooMany                = iota
	StatusNg                     = iota
)

type RingService[T any] struct {
	wreqh WriteRequestHandler
	wevth WroteEventHandler
	lrht  ListRequestHandler[T]
	drh   DeleteHandler[T]
	rh    ReadHandler[T]
}

type ServiceEvent struct {
	body   []byte
	stat   error
	status ServiceStatus
}

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

func (s RingService[T]) List(req ListRequest, wtr ListEventWriterTo[T]) ServiceEvent {
	l, e := s.lrht(req)
	return l.ToServiceEvent(e, wtr)
}

func (s RingService[T]) Write(req WriteRequest) ServiceEvent {
	f := ComposeErr(
		s.wreqh,
		func(evt WroteEvent) (WroteEvent, error) { return evt, s.wevth(evt) },
	)
	return ServiceEventNew(nil, ErrOnly(f)(req))
}

func (s RingService[T]) Del(req DeleteRequest[T]) ServiceEvent {
	evt, e := s.drh(req)
	return evt.ToServiceEvent(e)
}

func (s RingService[T]) Get(req ReadRequest[T]) ServiceEvent {
	r, e := s.rh(req)
	return r.ToServiceEvent(e)
}
