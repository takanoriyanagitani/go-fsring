package fsring

type App struct {
	wreqh WriteRequestHandler
	wevth WroteEventHandler
}

func (a App) HandleWriteRequest(req WriteRequest) error {
	f := ComposeErr(
		a.wreqh,
		func(evt WroteEvent) (WroteEvent, error) { return evt, a.wevth(evt) },
	)
	return ErrOnly(f)(req)
}
