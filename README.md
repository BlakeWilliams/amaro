# Amaro

Amaro is an experimental CLI framework for Go applications. It's intended to be extensible, so first and third party plugins can be added to extend the functionality of the CLI.

## Installation

```bash
go get github.com/blakewilliams/amaro
```

## Usage

```go
package main

import (
    "fmt"
    "os"

    "github.com/blakewilliams/amaro"
)

type serverCmd struct {
    Addr string `flag:"address"`
}

func main() {
    app := amaro.NewApplication("my app)
    app.RegisterCommand(&serverCmd{})
    app.Execute(context.Background())
}
```

Then, run your app:

```bash
go run main.go server --address localhost:3000
```
