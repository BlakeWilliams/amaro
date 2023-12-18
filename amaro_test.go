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

	app.RegisterCommand(&Greeter{}, "greet", "greets users")
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

	app.RegisterCommand(&AgeGreeter{}, "greet", "greets users")
	app.ExecuteWithArgs(context.Background(), []string{"greet", "--name", "Fox Mulder"})

	got := app.Out.(*bytes.Buffer).String()

	require.Contains(t, got, "missing required flag: age")
}

func TestCLI_IntArg(t *testing.T) {
	app := NewApplication("test")
	var b []byte
	app.Out = bytes.NewBuffer(b)

	app.RegisterCommand(&AgeGreeter{}, "greet", "greets users")
	app.ExecuteWithArgs(context.Background(), []string{"greet", "--name", "Fox Mulder", "--age=42"})

	got := app.Out.(*bytes.Buffer).String()
	expected := "Hello Fox Mulder, you are 42 years old!\n"

	require.Equal(t, expected, got)
}

func TestCLI__Help(t *testing.T) {
	app := NewApplication("test")
	var b []byte
	app.Out = bytes.NewBuffer(b)

	app.RegisterCommand(&AgeGreeter{}, "ageGreet", "greets users with age")
	app.ExecuteWithArgs(context.Background(), []string{"greet", "--name", "Fox Mulder"})

	app.RegisterCommand(&AgeGreeter{}, "greet", "greets users")
	app.ExecuteWithArgs(context.Background(), []string{"greet", "--name", "Fox Mulder"})

	got := app.Out.(*bytes.Buffer).String()

	require.Contains(t, got, "missing required flag: age")
}

var RegisterCommandTests = []struct {
	name     string
	runnable Runnable
	valid    bool
}{
	{"greet", &Greeter{}, true},
	{"ageGreet", &AgeGreeter{}, true},
	{"help", &AgeGreeter{}, false},
	{"wow omg", &AgeGreeter{}, false},
	{"wow-omg", &AgeGreeter{}, false},
	{"wow:omg", &AgeGreeter{}, true},
}

func TestRegisterCommand(t *testing.T) {
	for _, tt := range RegisterCommandTests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApplication("test")

			if tt.valid {
				require.NotPanics(t, func() {
					app.RegisterCommand(tt.runnable, tt.name, "desc")
				})
			} else {
				require.Panics(t, func() {
					app.RegisterCommand(tt.runnable, tt.name, "desc")
				})
			}
		})
	}
}

func TestHelp(t *testing.T) {

	app := NewApplication("test")
	var b []byte
	app.Out = bytes.NewBuffer(b)

	app.RegisterCommand(&Greeter{}, "greet", "greets users")
	app.RegisterCommand(&AgeGreeter{}, "greet:age", "greets users with their age")
	app.ExecuteWithArgs(context.Background(), []string{"help"})

	expected := "usage\n  greet      greets users\n  greet:age  greets users with their age\n"
	got := app.Out.(*bytes.Buffer).String()

	require.Equal(t, expected, got)
}

func TestHelpCommand(t *testing.T) {
	app := NewApplication("test")
	var b []byte
	app.Out = bytes.NewBuffer(b)

	app.RegisterCommand(&Greeter{}, "greet", "greets users")
	app.RegisterCommand(&AgeGreeter{}, "greet:age", "greets users with their age")
	app.ExecuteWithArgs(context.Background(), []string{"help", "greet:age"})

	expected := `usage for greet:age
  -name  The name of the person to greet
  -age   The age of the person to greet (required)
`
	got := app.Out.(*bytes.Buffer).String()

	require.Equal(t, expected, got)
}
