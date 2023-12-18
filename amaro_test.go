package amaro

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type Greeter struct {
	Name string `flag:"name" description:"The name of the person to greet"`
}

func (g *Greeter) RunCommand(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, "Hello %s!\n", g.Name)
	return nil
}

func TestCLICall(t *testing.T) {
	app := NewApplication("test")
	var b []byte
	app.Out = bytes.NewBuffer(b)

	app.RegisterCommand("greet", &Greeter{})
	app.ExecuteWithArgs(context.Background(), []string{"greet", "--name", "Fox Mulder"})

	expected := "Hello Fox Mulder!\n"
	got := app.Out.(*bytes.Buffer).String()

	require.Equal(t, expected, got)
}

type AgeGreeter struct {
	Name string `flag:"name" description:"The name of the person to greet"`
	Age  int    `flag:"age" required:"true" description:"The age of the person to greet"`
}

func (g *AgeGreeter) RunCommand(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, "Hello %s, you are %d years old!\n", g.Name, g.Age)
	return nil
}

func TestCLI_MissingRequiredArg(t *testing.T) {
	app := NewApplication("test")
	var b []byte
	app.Out = bytes.NewBuffer(b)

	app.RegisterCommand("greet", &AgeGreeter{})
	app.ExecuteWithArgs(context.Background(), []string{"greet", "--name", "Fox Mulder"})

	got := app.Out.(*bytes.Buffer).String()

	require.Contains(t, got, "missing required flag: age")
}
