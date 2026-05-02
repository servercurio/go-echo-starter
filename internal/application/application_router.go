package application

func (app *Application) initializeRouting() error {
	server := app.httpServer

	if app.tlsEnabled() {
		server = app.tlsServer
	}

	for _, m := range app.modules {
		g := server.Group(m.Prefix(), m.Middleware()...)
		m.AttachGroup(g)
	}

	return nil
}
