package vite

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Vite represents an instance of Vite that can be used to build assts for
// production or serve them in development.
type Vite struct {
	Port    int
	ready   chan struct{}
	running bool
}

type ViteOption func(*Vite)

// WithPort sets the port that Vite will use to serve assets.
func WithPort(port int) ViteOption {
	return func(v *Vite) {
		v.Port = port
	}
}

func New(opts ...ViteOption) *Vite {
	v := &Vite{
		Port:  5173,
		ready: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(v)
	}

	return v
}

func (v *Vite) Ready() <-chan struct{} {
	return v.ready
}

func (v *Vite) Middleware() func(w http.ResponseWriter, r *http.Request, next http.Handler) {
	url, err := url.Parse(fmt.Sprintf("http://localhost:%d", v.Port))
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(url)

	return func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		if strings.HasPrefix(r.URL.Path, "/assets") {
			proxy.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (v *Vite) Start(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "npx", "vite", "serve", "--port", strconv.Itoa(v.Port), "--strictPort")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	v.running = true
	go func() {
		for {
			res, err := http.Get(fmt.Sprintf("http://localhost:%d/", v.Port))
			if err != nil {
				if !v.running {
					break
				}
				continue
			}

			if res.StatusCode != 0 {
				close(v.ready)
			}

			time.Sleep(time.Millisecond * 25)
		}
	}()

	res := cmd.Run()
	v.running = false

	return res
}
