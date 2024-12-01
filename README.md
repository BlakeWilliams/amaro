# Amaro

If frameworks were operating systems, Rails would be Ubuntu and Amaro would be
Arch. Amaro attempts to include all the bells and whistles needed for
production apps, but the glue and application is up to you. (guided, of course)

Amaro is a composable application framework for Go, aiming to create a small
ecosystem of components like http routers, job runners, test helpers, and other
frameworks/tooling necessary to make complex Go applications.

## Getting Started

The core Amaro package (github.com/blakewilliams/amaro) is focused on creating
runnable commands in your project, and is how other packages hook into your
project.

For example, to get started with a web project you would create
`cmd/appname/main.go` and include the following:

```go
package main

import (
	"context"

	"github.com/blakewilliams/amaro"
    "github.com/you/project/core"
)

func main() {
	// Application must implement `AppName()` and `Log()`
	runner := amaro.NewApplication(&core.Application{
		// your application configuration
    })
	runner.Execute(context.TODO())
}
```

This requires a "core" application, that must implement the `amaro.Application` interface.

### Adding a command

Adding commands is simple, implement the `Command[T]` interface on a struct
(typically a dedicated struct):

```
type GreetCmd[T amaro.Application] struct {
	Name string
}

func (g *GreetCmd[T]) CommandName() string { return "greet" }
func (g *GreetCmd[T]) CommandDescription() string { return "Greets you" }
// The actual command
func (g *GreetCmd[T]) RunCommand(ctx context.Context, app T) {
	app.Log(fmt.Sprintf("Hello, %s!", g.Name))
}
```

Then register your command in your `main` function like
`runner.RegisterCommand(&GreetCmd{Name: "Fox Mulder"})` and run your command
with `go run <your cmd path> greet`!

## Web

TODO document how to bootstrap a web app
