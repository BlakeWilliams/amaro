package csrf

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blakewilliams/amaro/httprouter"
	"github.com/stretchr/testify/require"
)

type sessionData struct {
	CSRF *Token
}
type requestContext struct {
	session *sessionData

	httprouter.RequestContext
}

func (r *requestContext) SetSessionData(s *sessionData) { r.session = s }
func (r *requestContext) SessionData() *sessionData     { return r.session }
func (r *requestContext) CSRF() *Token                  { return r.SessionData().CSRF }
func (r *requestContext) SetCSRF(t *Token)              { r.SessionData().CSRF = t }

func TestMiddleware(t *testing.T) {
	t.Run("creates and sets token with the provided length", func(t *testing.T) {
		rctx := newRequestContext(http.MethodGet)
		called := false
		next := func(ctx context.Context, rctx *requestContext) {
			called = true
		}

		Middleware(MiddlewareConfig[*requestContext]{
			TokenLength: 16,
		})(context.Background(), rctx, next)

		require.True(t, called)
		require.Len(t, rctx.CSRF().Value, 16)
	})

	t.Run("does not overwrite existing tokens", func(t *testing.T) {
		rctx := newRequestContext(http.MethodGet)
		called := false
		next := func(ctx context.Context, rctx *requestContext) {
			called = true
		}

		Middleware(MiddlewareConfig[*requestContext]{
			TokenLength: 16,
		})(context.Background(), rctx, next)

		existingToken := rctx.CSRF().Value

		Middleware(MiddlewareConfig[*requestContext]{
			TokenLength: 16,
		})(context.Background(), rctx, next)

		require.True(t, called)
		require.Equal(t, existingToken, rctx.CSRF().Value)
	})

	t.Run("calls provided error handler when token is invalid", func(t *testing.T) {
		rctx := newRequestContext(http.MethodPost)
		rctx.SetCSRF(&Token{TokenLength: 10, Value: []byte("invalid woo")})
		called := false
		next := func(ctx context.Context, rctx *requestContext) {
			called = true
		}

		invalidHandlerCalled := false
		Middleware(MiddlewareConfig[*requestContext]{
			TokenLength: 16,
			HandleInvalidToken: func(ctx context.Context, rctx *requestContext) {
				invalidHandlerCalled = true
			},
		})(context.Background(), rctx, next)

		require.False(t, called)
		require.True(t, invalidHandlerCalled)
	})

	t.Run("accepts CSRFs in header", func(t *testing.T) {
		rctx := newRequestContext(http.MethodPost)
		rctx.SetCSRF(NewCSRF(WithTokenLength(16)))
		rctx.Request().Header.Set("X-CSRF-Token", rctx.CSRF().AuthenticityToken())
		called := false
		next := func(ctx context.Context, rctx *requestContext) {
			called = true
		}

		Middleware(MiddlewareConfig[*requestContext]{
			TokenLength: 16,
		})(context.Background(), rctx, next)

		require.True(t, called)
	})
}

func newRequestContext(method string) *requestContext {
	req := httptest.NewRequest(method, "/", nil)
	res := httptest.NewRecorder()
	rootRctx := httprouter.NewRequestContext(req, res, "/", map[string]string{})
	return &requestContext{
		session:        &sessionData{},
		RequestContext: rootRctx,
	}
}
