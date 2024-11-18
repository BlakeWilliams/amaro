package generator

func generateWeb(d *driver) error {
	templateData := map[string]any{
		"PackageName": d.packageName,
	}

	d.createFile("internal/web/server.go", serverTemplate, templateData)
	d.createFile("internal/web/routes.go", `package web

func regiterRoutes(s *Server) {
	s.router.Get("/", homeHandler)
}
`, templateData)

	d.createFile("internal/web/request_context.go", requestContextTemplate, templateData)
	d.createFile("internal/web/site_handlers.go", siteHandlersTemplate, templateData)
	d.createFile("internal/web/home_handler_test.go", homeHandlerTest, templateData)
	d.createFile("internal/web/components/components.go", componentsTemplate, templateData)
	d.createFile("internal/web/components/errors.go", errorsTemplate, templateData)
	d.createFile("internal/web/components/templates/main_layout.html", mainLayoutTemplate, templateData)
	d.createFile("internal/web/components/templates/site/home.html", homeTemplate, templateData)
	d.createFile("internal/web/components/templates/errors/500.html", errorsTemplateMarkup, templateData)

	return nil
}

const serverTemplate = `package web

import (
	"context"
	"net/http"
	"github.com/blakewilliams/amaro/httprouter"
	"github.com/blakewilliams/amaro/httprouter/metal"
	"github.com/blakewilliams/amaro/httprouter/middleware"
	"{{.PackageName}}/internal/web/components"
	"github.com/blakewilliams/glam"
	"{{.PackageName}}/internal/core"
)


// Server represents the web server and routes of the application.
type Server struct {
	router   *httprouter.Router[*requestContext]
	app      *core.Application
	renderer *glam.Engine
}

func NewServer(app *core.Application) *Server {
	s := &Server{app: app}
	s.router = initRouter(s)
	s.renderer = components.New(nil)

	return s
}

func initRouter(s *Server) *httprouter.Router[*requestContext] {
	r := httprouter.New[*requestContext](newRequestContext(s))
	r.UseMetal(metal.MethodRewrite)
	r.Use(middleware.ErrorHandler(s.app.Logger, errorHandler))

	r.Get("/", homeHandler)

	return r
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
`

const requestContextTemplate = `package web

import (
	"bytes"
	"context"
	"io"
	"github.com/blakewilliams/amaro/httprouter"
	"github.com/blakewilliams/glam"
	"{{.PackageName}}/internal/web/components"
	"{{.PackageName}}/internal/core"
)

// requestContext holds the data needed for the application to render a response
// to an http request. It embeds the router.RequestContext type, which holds
// the http.Request and http.ResponseWriter.
type requestContext struct {
	Title string
	renderer *glam.Engine
	httprouter.RequestContext
	Environment core.Env
}

// newRequestContext returns a new RequestContext factory that router calls for
// each route and passes the resulting RequestContext to the handler function.
func newRequestContext(s *Server) func(r httprouter.RequestContext) *requestContext {
	return func(r httprouter.RequestContext) *requestContext {
		return &requestContext{
			renderer:          s.renderer,
			RequestContext:	   r,
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
		Children:    b.String(),
		Title:       rc.Title,
	})
}
`

const siteHandlersTemplate = `package web

import (
	"context"
	"{{.PackageName}}/internal/web/components"
)

// homeHandler renders the home page of the site.
func homeHandler(ctx context.Context, rc *requestContext) {
	rc.Render(ctx, components.Home{Message: "Hello, amaro!"})
}`

const componentsTemplate = `package components
import (
	"embed"
	"html/template"

	"github.com/blakewilliams/glam"
)

// requestFuncMap is a map of functions that are request specific and are
// passed to the template at render time. See the requestContext.Render method.
var requestFuncMap = glam.FuncMap{
	"CSRFInput": func() template.HTML {
		panic("pass in render")
	},
}

//go:embed all:templates/*
var templateFS embed.FS

func New(funcMap glam.FuncMap) *glam.Engine {
	funcs := make(glam.FuncMap, len(requestFuncMap)+len(funcMap))
	for k, v := range requestFuncMap {
		funcs[k] = v
	}
	for k, v := range funcMap {
		funcs[k] = v
	}

	e := glam.New(funcs)
	err := e.RegisterManyFS(templateFS, map[any]string{
		&MainLayout{}:   "templates/main_layout.html",
		&Home{}:   "templates/site/home.html",
		&Err500{}:   "templates/errors/500.html",
	})

	if err != nil {
		panic(err)
	}

	return e
}

type MainLayout struct {
	Title        string
	Children	 string
}

func (l MainLayout) TitleTag() template.HTML {
	title := "{{.PackageName}}"
	if (l.Title != "") {
		title = template.HTMLEscapeString(l.Title)
	}

	return template.HTML("<title>" + title + "</title>")
}

type Home struct {
	Message string
}`

const errorsTemplate = `package components

import (
	"runtime/debug"
)

import (
	"{{.PackageName}}/internal/core"
)

type Err500 struct {
	Environment core.Env
	Error any
}

func (e *Err500) renderDetailedError() bool {
	return e.Environment != "production"
}

func (e *Err500) errorType() string {
	if err, ok := e.Error.(error); ok {
		return err.(error).Error()
	} else {
		return "Unknown error type"
	}
}

func (e *Err500) errorDetails() string {
	return string(debug.Stack())
}
`

const errorsTemplateMarkup = `
<h1>Oops, something went wrong!</h1>

{{"{{ if .renderDetailedError }}"}}
  <pre>{{"{{.errorType}}"}}</pre>
  <pre>{{"{{.errorDetails}}"}}</pre>
{{"{{ end }}"}}
`

const mainLayoutTemplate = `<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="UTF-8" />
  <meta http-equiv="X-UA-Compatible" content="ie=edge" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  {{"{{.TitleTag}}"}}
</head>

<body class="dark:bg-gray-950 text-gray-900 dark:text-gray-100">
  <h1>{{"{{.Children}}"}}</h1>
</body>
</html>`

const homeTemplate = `<h1>{{"{{.Message}}"}}</h1>`

const homeHandlerTest = `package web

import (
	"log/slog"
	"io"
	"testing"
	"{{.PackageName}}/internal/core"
	"github.com/blakewilliams/amaro/apptest"
	"github.com/stretchr/testify/require"
)

func startApp(t testing.TB) (*Server, func()) {
	app := &core.Application{
		Environment: core.EnvTest,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

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
	require.Contains(t, res.BodyString(), "Hello, amaro")
}
`
