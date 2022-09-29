package fsring

import (
	"fmt"
)

type App struct {
	wreqh WriteRequestHandler
	wevth WroteEventHandler
}

func AppNew(wreqh WriteRequestHandler, wevth WroteEventHandler) (App, error) {
	var valid bool = nil != wreqh && nil != wevth
	return ErrFromBool(
		valid,
		func() App { return App{wreqh, wevth} },
		func() error { return fmt.Errorf("Invalid arguments") },
	)
}

func (a App) HandleWriteRequest(req WriteRequest) error {
	f := ComposeErr(
		a.wreqh,
		func(evt WroteEvent) (WroteEvent, error) { return evt, a.wevth(evt) },
	)
	return ErrOnly(f)(req)
}
