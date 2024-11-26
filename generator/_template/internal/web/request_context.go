package web

import (
	"bytes"
	"context"
	"html/template"
	"io"

	"github.com/blakewilliams/amaro/_template/internal/core"
	"github.com/blakewilliams/amaro/_template/internal/web/components"
	"github.com/blakewilliams/amaro/httprouter"
	"github.com/blakewilliams/glam"
)

// requestContext holds the data needed for the application to render a response
// to an http request. It embeds the router.RequestContext type, which holds
// the http.Request and http.ResponseWriter.
type requestContext struct {
	Title    string
	renderer *glam.Engine
	httprouter.RequestContext
	sessionData *sessionData
	Environment core.Env
}

// newRequestContext returns a new RequestContext factory that router calls for
// each route and passes the resulting RequestContext to the handler function.
func newRequestContext(s *Server) func(r httprouter.RequestContext) *requestContext {
	return func(r httprouter.RequestContext) *requestContext {
		return &requestContext{
			renderer:       s.renderer,
			RequestContext: r,
		}
	}
}

// RenderTo renders the given component with the provided data into the provided
// io.Writer.
func (rc *requestContext) RenderTo(ctx context.Context, w io.Writer, component any) {
	err := rc.renderer.RenderWithFuncs(w, component, glam.FuncMap{
		// "CSRFToken": func() string {
		// 	return rc.SessionData().CSRF.AuthenticityToken()
		// },
	})

	if err != nil {
		panic(err)
	}
}

// Render renders the given template with the provided data. The data is merged
// with the default render data, and the result is written to the response.
func (rc *requestContext) Render(ctx context.Context, component any) {
	var b bytes.Buffer
	rc.RenderTo(
		ctx,
		&b,
		component,
	)

	rc.RenderTo(ctx, rc.Response(), &components.MainLayout{
		Children: template.HTML(b.String()),
		Title:    rc.Title,
	})
}

// SetSessionData implements part of the session.Persistable interface so that
// the session data is accessible during the request.
func (rc *requestContext) SetSessionData(s *sessionData) {
	rc.sessionData = s
}

// SessionData implements part of the session.Persistable interface so that
// the session data can be accessed by the application and the session package.
func (rc *requestContext) SessionData() *sessionData {
	return rc.sessionData
}
