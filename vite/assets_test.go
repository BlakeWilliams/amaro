package vite

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/blakewilliams/amaro/httprouter"
	"github.com/stretchr/testify/require"
)

var server *Vite

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	server = New()
	go func() {
		err := server.Start(ctx)
		if err != nil {
			panic(err)
		}
	}()

	select {
	case <-server.Ready():
		break
	case <-time.After(time.Second * 5):
		panic("could not start vite server")
	}

	if !server.running {
		panic("could not start vite server")
	}

	res := m.Run()
	cancel()
	os.Exit(res)
}

func TestVite(t *testing.T) {
	v := New()
	if v.Port != 5173 {
		require.Equal(t, 5173, v.Port)
	}

	res, err := http.Get("http://localhost:5173/assets/main.js")
	require.NoError(t, err)

	require.Equal(t, 200, res.StatusCode)
	body, err := io.ReadAll(res.Body)

	require.Contains(t, string(body), "I want to believe")
}

func TestMiddleware(t *testing.T) {
	v := New()
	if v.Port != 5173 {
		require.Equal(t, 5173, v.Port)
	}

	router := httprouter.New(func(h httprouter.RequestContext) httprouter.RequestContext {
		return h
	})
	router.UseMetal(v.Middleware())

	server := httptest.NewServer(router)
	defer server.Close()

	res, err := http.Get(fmt.Sprintf("http://localhost:%d/assets/index.js", v.Port))
	require.NoError(t, err)

	require.Equal(t, 200, res.StatusCode)
	body, err := io.ReadAll(res.Body)

	require.Contains(t, string(body), "I want to believe")
}
