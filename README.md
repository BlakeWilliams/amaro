# Amaro

Amaro is a composable application framework for Go, aiming to create a small ecosystem of components like http routers, job runners, test helpers, and other frameworks/tooling necessary to make complex Go applications.

## Getting Started

The core Amaro package (github.com/blakewilliams/amaro) is focused on creating runnable commands in your project, and is how other packages hook into your project.

For example, to get started with a web project you would create `cmd/appname/main.go` and include the following:

```go
package main

import (
	"context"

	"github.com/blakewilliams/amaro"
	"github.com/blakewilliams/amaro/_template/internal/web"
)

func main() {
	runner := amaro.NewApplication("myapp")
	runner.RegisterCommand(&web.Command{
		Addr: ":8080",
	})
	runner.Execute(context.TODO())
}
```

Then you can run `go run cmd/appname/main.go generate:core` to generate the base files of the project.
