package session

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blakewilliams/amaro/httprouter"
	"github.com/stretchr/testify/require"
)

type MyData struct {
	UserID int
	Name   string
}

type requestContext struct {
	session *MyData
	httprouter.RequestContext
}

func (rc *requestContext) SetSessionData(session *MyData) {
	rc.session = session
}

func (rc *requestContext) SessionData() *MyData {
	return rc.session
}

func TestMiddleware(t *testing.T) {
	verifier := NewVerifier("TheTruthIsOutThere")
	router := httprouter.New(func(rc httprouter.RequestContext) *requestContext {
		return &requestContext{
			RequestContext: rc,
		}
	})

	// TODO: remove init function and rely on SessionData() to initialize
	// session.
	store := New[*MyData]("session", verifier, nil, func() *MyData { return &MyData{} })
	cookie, err := store.cookie(&MyData{UserID: 500, Name: "Fox Mulder"})
	require.NoError(t, err)

	router.Use(Middleware[*requestContext, *MyData](store))

	router.Get("/", func(ctx context.Context, rc *requestContext) {
		require.Equal(t, 500, rc.session.UserID)
		require.Equal(t, "Fox Mulder", rc.session.Name)

		rc.session.UserID = 200
		rc.session.Name = "Dana Scully"
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	newCookie := res.Result().Cookies()[0]
	data, err := store.FromCookie(newCookie)
	require.NoError(t, err)

	require.Equal(t, 200, data.UserID)
	require.Equal(t, "Dana Scully", data.Name)

	// Safe defaults
	require.Equal(t, http.SameSiteLaxMode, newCookie.SameSite)
	require.True(t, newCookie.Secure)
	require.True(t, newCookie.HttpOnly)
}

func TestMiddleware_Init(t *testing.T) {
	verifier := NewVerifier("TheTruthIsOutThere")
	router := httprouter.New(func(rc httprouter.RequestContext) *requestContext {
		return &requestContext{
			RequestContext: rc,
		}
	})

	store := New[*MyData]("session", verifier, nil, func() *MyData { return &MyData{} })
	router.Use(Middleware[*requestContext, *MyData](store))

	var innerRC *requestContext
	router.Get("/", func(ctx context.Context, rc *requestContext) {
		require.Equal(t, 0, rc.session.UserID)
		require.Equal(t, "", rc.session.Name)

		rc.session.UserID = 500
		rc.session.Name = "Fox Mulder"

		innerRC = rc
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	require.Equal(t, 500, innerRC.session.UserID)
	require.Equal(t, "Fox Mulder", innerRC.session.Name)
}

func TestCookieOptions(t *testing.T) {
	verifier := NewVerifier("TheTruthIsOutThere")
	options := &CookieOptions{
		Domain:   "example.com",
		MaxAge:   3600,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
		HTTPOnly: true,
	}
	store := New[*MyData]("session", verifier, options, func() *MyData { return &MyData{} })

	cookie, err := store.cookie(&MyData{UserID: 500, Name: "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "example.com", cookie.Domain)
	require.Equal(t, 3600, cookie.MaxAge)
	require.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	require.True(t, cookie.Secure)
	require.True(t, cookie.HttpOnly)
}
