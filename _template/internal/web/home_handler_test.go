package web

import (
	"io"
	"log/slog"
	"testing"

	"github.com/blakewilliams/amaro/_template/internal/core"
	"github.com/blakewilliams/amaro/apptest"
	"github.com/stretchr/testify/require"
)

func startApp(t testing.TB) (*Server, func()) {
	app := core.New(core.EnvTest)
	app.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	server := NewServer(app)

	return server, func() {
		// TODO add cleanup here, like DB stopping services
	}
}

func TestHomeHandler(t *testing.T) {
	app, stop := startApp(t)
	defer stop()

	session := apptest.New(app)
	res := session.Get("/", nil)

	require.Equal(t, 200, res.Code())
	require.Contains(t, res.Body(), "Hello, amaro")
}
