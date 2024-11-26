package main

import (
	"context"

	"github.com/blakewilliams/amaro"
	"github.com/blakewilliams/amaro/_template/internal/web"
)

func main() {
	runner := amaro.NewApplication("_placeholder_")
	runner.RegisterCommand(&web.Command{
		Addr: ":8080",
	})
	runner.Execute(context.TODO())
}
