package csrf_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/blakewilliams/amaro/httprouter"
	"github.com/blakewilliams/amaro/httprouter/middleware"
	"github.com/blakewilliams/amaro/httprouter/middleware/session"
	"github.com/blakewilliams/amaro/httprouter/middleware/session/csrf"
	"github.com/stretchr/testify/require"
)

type sessionData struct {
	CSRF *csrf.Token
}
type requestContext struct {
	session *sessionData

	httprouter.RequestContext
}

func (r *requestContext) SetSessionData(s *sessionData) { r.session = s }
func (r *requestContext) SessionData() *sessionData     { return r.session }
func (r *requestContext) CSRF() *csrf.Token             { return r.SessionData().CSRF }
func (r *requestContext) SetCSRF(t *csrf.Token)         { r.SessionData().CSRF = t }

func Test(t *testing.T) {
	testCases := []struct {
		desc     string
		requests []string
	}{
		{
			desc:     "sets token in session and accepts in post",
			requests: []string{"GET / 200", "POST / 200"},
		},
		{
			desc:     "post with no session 500s",
			requests: []string{"POST / 500"},
		},
	}
	for _, tC := range testCases {
		r := httprouter.New(func(rctx httprouter.RequestContext) *requestContext {
			return &requestContext{
				RequestContext: rctx,
			}
		})

		verifier := session.NewVerifier("TheTruthIsOutThere")
		store := session.New("session", verifier, nil, func() *sessionData { return &sessionData{} })

		logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
		r.Use(middleware.ErrorHandler[*requestContext](logger, func(ctx context.Context, rctx *requestContext, recovered any) {
			rctx.Response().WriteHeader(http.StatusInternalServerError)
			rctx.Response().Write([]byte("500"))
		}))

		r.Use(session.Middleware[*requestContext](store))
		r.Use(csrf.Middleware(csrf.MiddlewareConfig[*requestContext]{}))

		r.Get("/", func(ctx context.Context, rctx *requestContext) {
			rctx.Response().Write([]byte(
				rctx.CSRF().AuthenticityToken(),
			))
		})
		r.Post("/", func(ctx context.Context, rctx *requestContext) {
			rctx.Response().Write([]byte(
				rctx.CSRF().AuthenticityToken(),
			))
		})

		t.Run(tC.desc, func(t *testing.T) {
			cookies := []string{}
			csrf := ""

			for _, req := range tC.requests {
				parts := strings.Split(req, " ")
				method := parts[0]
				route := parts[1]
				expectedStatus := parts[2]

				req := httptest.NewRequest(method, route, nil)
				if method != "GET" {
					reqBody := url.Values{"authenticity_token": {csrf}}
					req = httptest.NewRequest(method, route, strings.NewReader(reqBody.Encode()))
					req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				}
				for _, cookie := range cookies {
					req.Header.Add("Cookie", cookie)
				}
				res := httptest.NewRecorder()

				r.ServeHTTP(res, req)
				// set cookies for next request
				cookies = res.Header().Values("Set-Cookie")

				require.Equal(
					t,
					expectedStatus,
					strconv.Itoa(res.Code),
					fmt.Sprintf("for %s %s", method, route),
				)

				if newCSRF := res.Body.String(); newCSRF != "" {
					csrf = newCSRF
					fmt.Println(csrf)
				}
			}
		})
	}
}
