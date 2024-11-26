package web

import (
	"context"
	"net/http"
	"os"

	"github.com/blakewilliams/amaro/_template/internal/core"
	"github.com/blakewilliams/amaro/_template/internal/web/components"
	"github.com/blakewilliams/amaro/httprouter"
	"github.com/blakewilliams/amaro/httprouter/metal"
	"github.com/blakewilliams/amaro/httprouter/middleware"
	"github.com/blakewilliams/amaro/httprouter/middleware/session"
	"github.com/blakewilliams/glam"
)

type sessionData struct {
	// TODO handle flash
}

// Server represents the web server and routes of the application.
type Server struct {
	router       *httprouter.Router[*requestContext]
	app          *core.Application
	renderer     *glam.Engine
	sessionStore session.Store[*sessionData]
}

func NewServer(app *core.Application) *Server {
	s := &Server{app: app}
	s.sessionStore = initSessionStore(s)
	s.renderer = components.New(nil)
	s.router = initRouter(s)

	/// see web/routes.go
	s.registerRoutes()

	return s
}

func initRouter(s *Server) *httprouter.Router[*requestContext] {
	r := httprouter.New[*requestContext](newRequestContext(s))
	r.UseMetal(metal.MethodRewrite)
	r.Use(middleware.ErrorHandler(s.app.Logger, errorHandler))
	r.Use(session.Middleware[*requestContext, *sessionData](s.sessionStore))

	return r
}

func initSessionStore(s *Server) session.Store[*sessionData] {
	cookieOpts := &session.CookieOptions{
		HTTPOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	verifierKey := os.Getenv("APPLICATION_VERIFIER_SECRET")
	if verifierKey == "" {
		panic("APPLICATION_VERIFIER_SECRET must be set")
	}
	verifier := session.NewEncryptedVerifier(verifierKey)

	return session.New("github.com/blakewilliams/amaro/_template_session", verifier, cookieOpts, func() *sessionData {
		return &sessionData{}
	})
}

func errorHandler(ctx context.Context, rc *requestContext, r any) {
	if rc.Environment == "prod" {
		rc.Response().WriteHeader(http.StatusInternalServerError)
		rc.Render(ctx, components.Err500{Environment: rc.Environment, Error: r})
		return
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
