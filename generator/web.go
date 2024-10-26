package generator

func generateWeb(d *driver) error {
	templateData := map[string]any{
		"PackageName": d.packageName,
	}

	// Create the server
	d.createFile("internal/web/server.go", serverTemplate, templateData)

	// Create the routes file
	d.createFile("internal/web/routes.go", `package web

func regiterRoutes(s *Server) {
	s.router.Get("/", homeHandler)
}
`, templateData)

	d.createFile("internal/web/request_context.go", requestContextTemplate, templateData)
	d.createFile("internal/web/site_handlers.go", siteHandlersTemplate, templateData)
	d.createFile("internal/web/components/components.go", componentsTemplate, templateData)
	d.createFile("internal/web/components/templates/main_layout.html", mainLayoutTemplate, templateData)
	d.createFile("internal/web/components/templates/site/home.html", homeTemplate, templateData)

	return nil
}

const serverTemplate = `package web

import (
	"github.com/blakewilliams/amaro/httprouter"
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
	s.router = httprouter.New[*requestContext](newRequestContext(s))

	return s
}

func initRouter(s *Server) *httprouter.Router[*requestContext] {
	r := httprouter.New[*requestContext](newRequestContext(s))

	r.Get("/", homeHandler)

	return r
}`

const requestContextTemplate = `package web

import (
	"bytes"
	"context"
	"io"
	"github.com/blakewilliams/amaro/httprouter"
	"github.com/blakewilliams/glam"
	"{{.PackageName}}/internal/web/components"
)

// requestContext holds the data needed for the application to render a response
// to an http request. It embeds the router.RequestContext type, which holds
// the http.Request and http.ResponseWriter.
type requestContext struct {
	Title string
	renderer *glam.Engine
	httprouter.RequestContext
}

// newRequestContext returns a new RequestContext factory that router calls for
// each route and passes the resulting RequestContext to the handler function.
func newRequestContext(s *Server) func(r httprouter.RequestContext) *requestContext {
	return func(r httprouter.RequestContext) *requestContext {
		return &requestContext{
			renderer:          s.renderer,
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
		&MainLayout{}:   "templates/layout.html",
		&Home{}:   "templates/layout.html",
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

const mainLayoutTemplate = `<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="UTF-8" />
  <meta http-equiv="X-UA-Compatible" content="ie=edge" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <script src="{{ .MainJSPath }}"></script>
  {{"{{.TitleTag}}"}}
</head>

<body class="dark:bg-gray-950 text-gray-900 dark:text-gray-100">
  <h1>{{"{{children}}"}}</h1>
</body>
</html>`

const homeTemplate = `<h1>{{"{{Message}}"}}</h1>`
