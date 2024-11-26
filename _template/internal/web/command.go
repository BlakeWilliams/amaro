package web

import (
	"context"
	"io"
	"net/http"

	"github.com/blakewilliams/amaro/_template/internal/core"
)

type Command struct {
	Addr string
}

func (s *Command) CommandName() string {
	return "serve"
}

func (s *Command) CommandDescription() string {
	return "Starts the web server"
}

func (s *Command) RunCommand(ctx context.Context, w io.Writer) error {
	app := core.New(core.EnvDevelopment)
	server := NewServer(app)

	done := make(chan struct{})
	app.Logger.Info("starting server", "addr", s.Addr)
	go func() {
		defer close(done)
		if err := http.ListenAndServe(s.Addr, server); err != http.ErrServerClosed {
			app.Logger.Error("error starting server")
		}
	}()

	select {
	case <-ctx.Done():
	case <-done:
	}

	return nil
}
