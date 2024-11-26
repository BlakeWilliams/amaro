package components

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
		&MainLayout{}: "templates/main_layout.html",
		&Home{}:       "templates/site/home.html",
		&Err500{}:     "templates/errors/500.html",
	})

	if err != nil {
		panic(err)
	}

	return e
}

type MainLayout struct {
	Title    string
	Children template.HTML
}

func (l MainLayout) TitleTag() template.HTML {
	title := "github.com/testing/testing"
	if l.Title != "" {
		title = template.HTMLEscapeString(l.Title)
	}

	return template.HTML("<title>" + title + "</title>")
}

type Home struct {
	Message string
}

