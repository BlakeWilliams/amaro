package flash_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blakewilliams/amaro/httprouter"
	"github.com/blakewilliams/amaro/httprouter/middleware/session"
	"github.com/blakewilliams/amaro/httprouter/middleware/session/flash"
	"github.com/stretchr/testify/require"
)

type sessionData struct {
	Flash *flash.Messages
}

type requestContext struct {
	sessionData *sessionData
	httprouter.RequestContext
}

func (r *requestContext) Flash() *flash.Messages {
	return r.sessionData.Flash
}

func (r *requestContext) SessionData() *sessionData {
	return r.sessionData
}

func (r *requestContext) SetSessionData(sd *sessionData) {
	r.sessionData = sd
}

func TestFlash(t *testing.T) {
	store := session.New("testing", session.NewVerifier("iiiiiiiiiiiiiiii"), nil, func() *sessionData {
		return &sessionData{
			Flash: &flash.Messages{},
		}
	})
	router := httprouter.New(func(rctx httprouter.RequestContext) *requestContext {
		return &requestContext{
			RequestContext: rctx,
		}
	})

	router.Use(session.Middleware[*requestContext](store))
	router.Use(flash.Middleware)

	router.Get("/", func(ctx context.Context, rctx *requestContext) {
		rctx.Flash().Set("foo", "hello!")
	})

	router.Get("/flash", func(ctx context.Context, rctx *requestContext) {
		rctx.Response().Write(
			[]byte(rctx.Flash().Get("foo")),
		)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	cookies := res.Header().Get("Set-Cookie")
	require.NotEmpty(t, cookies)

	req = httptest.NewRequest(http.MethodGet, "/flash", nil)
	req.Header.Set("Cookie", cookies)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, "hello!", res.Body.String())
}
