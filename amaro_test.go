package amaro

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type testApp struct {
	Name string
	// used to test command output
	out io.Writer
}

func (a *testApp) AppName() string {
	return a.Name
}

func (a *testApp) Log(msg string) {
	out := a.out
	if a.out == nil {
		out = os.Stdout
	}

	fmt.Fprint(out, msg)
}

func (a *testApp) Port() int {
	return 8080
}

type Greeter[T Application] struct {
	Name string `flag:"name" description:"The name of the person to greet"`
}

func (g *Greeter[T]) RunCommand(ctx context.Context, a T) error {
	a.Log(fmt.Sprintf("Hello %s!\n", g.Name))
	return nil
}

func (g *Greeter[T]) CommandName() string {
	return "greet"
}

func (g *Greeter[T]) CommandDescription() string {
	return "greets users"
}

func TestCLICall(t *testing.T) {
	var b []byte
	runner := NewApplication(&testApp{Name: "test", out: bytes.NewBuffer(b)})

	runner.RegisterCommand(&Greeter[*testApp]{})
	runner.ExecuteWithArgs(context.Background(), []string{"greet", "--name", "Fox Mulder"})

	expected := "Hello Fox Mulder!\n"
	got := runner.app.out.(*bytes.Buffer).String()

	require.Equal(t, expected, got)
}

type AgeGreeter[T Application] struct {
	Name string `flag:"name" description:"The name of the person to greet"`
	Age  int    `flag:"age" required:"true" description:"The age of the person to greet"`
}

func (g *AgeGreeter[T]) RunCommand(ctx context.Context, app T) error {
	app.Log(fmt.Sprintf("Hello %s, you are %d years old!\n", g.Name, g.Age))
	return nil
}

func (g *AgeGreeter[T]) CommandName() string {
	return "greet:age"
}

func (g *AgeGreeter[T]) CommandDescription() string {
	return "greets users with their age"
}

func TestCLI_MissingRequiredArg(t *testing.T) {
	var b []byte
	runner := NewApplication(&testApp{Name: "test", out: bytes.NewBuffer(b)})

	runner.RegisterCommand(&AgeGreeter[*testApp]{})
	runner.ExecuteWithArgs(context.Background(), []string{"greet:age", "--name", "Fox Mulder"})

	got := runner.app.out.(*bytes.Buffer).String()

	require.Contains(t, got, "missing required flag: age")
}

func TestCLI_IntArg(t *testing.T) {
	var b []byte
	runner := NewApplication(&testApp{Name: "test", out: bytes.NewBuffer(b)})

	runner.RegisterCommand(&AgeGreeter[*testApp]{})
	runner.ExecuteWithArgs(context.Background(), []string{"greet:age", "--name", "Fox Mulder", "--age=42"})

	got := runner.app.out.(*bytes.Buffer).String()
	expected := "Hello Fox Mulder, you are 42 years old!\n"

	require.Equal(t, expected, got)
}

func TestCLI__Help(t *testing.T) {
	var b []byte
	runner := NewApplication(&testApp{Name: "test", out: bytes.NewBuffer(b)})

	runner.RegisterCommand(&AgeGreeter[*testApp]{})
	runner.ExecuteWithArgs(context.Background(), []string{"greet:age", "--name", "Fox Mulder"})

	runner.RegisterCommand(&AgeGreeter[*testApp]{})
	runner.ExecuteWithArgs(context.Background(), []string{"greet:age", "--name", "Fox Mulder"})

	got := runner.app.out.(*bytes.Buffer).String()

	require.Contains(t, got, "missing required flag: age")
}

var RegisterCommandTests = []struct {
	name     string
	runnable Command[*testApp]
	valid    bool
}{
	{"greet", &Greeter[*testApp]{}, true},
	{"ageGreet", &AgeGreeter[*testApp]{}, true},
	{"help", &AgeGreeter[*testApp]{}, false},
	{"wow omg", &AgeGreeter[*testApp]{}, false},
	{"wow-omg", &AgeGreeter[*testApp]{}, false},
	{"wow:omg", &AgeGreeter[*testApp]{}, true},
}

func TestRegisterCommand(t *testing.T) {
	for _, tt := range RegisterCommandTests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewApplication(&testApp{Name: "test"})

			if tt.valid {
				require.NotPanics(t, func() {
					runner.RegisterCommandWithName(tt.runnable, tt.name)
				})
			} else {
				require.Panics(t, func() {
					runner.RegisterCommandWithName(tt.runnable, tt.name)
				})
			}
		})
	}
}

func TestHelp(t *testing.T) {
	var b []byte
	runner := NewApplication(&testApp{Name: "test", out: bytes.NewBuffer(b)})

	runner.RegisterCommand(&Greeter[*testApp]{})
	runner.RegisterCommand(&AgeGreeter[*testApp]{})
	runner.ExecuteWithArgs(context.Background(), []string{"help"})

	expected := "usage\n  greet      greets users\n  greet:age  greets users with their age\n"
	got := runner.app.out.(*bytes.Buffer).String()

	require.Contains(t, got, expected)
}

func TestHelpCommand(t *testing.T) {
	var b []byte
	runner := NewApplication(&testApp{Name: "test", out: bytes.NewBuffer(b)})

	runner.RegisterCommand(&Greeter[*testApp]{})
	runner.RegisterCommand(&AgeGreeter[*testApp]{})
	runner.ExecuteWithArgs(context.Background(), []string{"help", "greet:age"})

	expected := `usage for greet:age
  -name  The name of the person to greet
  -age   The age of the person to greet (required)
`
	got := runner.app.out.(*bytes.Buffer).String()

	require.Equal(t, expected, got)
}
