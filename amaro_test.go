package amaro

import (
	"bytes"
	"context"
	"fmt"
	"testing"
)

type Greeter struct {
	Name string `flag:"name" description:"The name of the person to greet"`
}

func (g *Greeter) RunCommand(ctx context.Context) error {
	fmt.Printf("Hello %s!\n", g.Name)
	return nil
}

func Test(t *testing.T) {
	app := NewApplication("test")
	var b []byte
	app.Out = bytes.NewBuffer(b)

	app.Register("greet", &Greeter{})
	app.ExecuteWithArgs([]string{"greet", "--name", "Fox Mulder!"})

	// expected := "Hello !\n"
}
