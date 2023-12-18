// Amaro is an application framework in the form of a library that integrates
// with your application. It's extensible so third-party packages can add
// functionality to your application in addition to first-party features added
// per-project.
package amaro

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type (
	Application struct {
		// The name of the application
		Name string

		// Out is the writer to which output is written. If nil, os.Stdout is used.
		Out io.Writer

		runnables            map[string]Runnable
		runnableDescriptions map[string]string
	}

	// Runnable is an interface that can be implemented by any type that
	// wants to be run by the application.
	Runnable interface {
		RunCommand(context.Context, io.Writer) error
	}
)

// NewApplication creates a new application instance.
func NewApplication(name string) *Application {
	return &Application{
		Name:                 name,
		runnables:            make(map[string]Runnable, 0),
		runnableDescriptions: make(map[string]string, 0),
	}
}

// Execute runs the registered runnable that matches the command line arguments.
// If no runnable matches, the help text is printed.
func (a *Application) Execute(ctx context.Context) {
	args := os.Args[1:]

	a.ExecuteWithArgs(ctx, args)
}

// ExecuteWithArgs runs the registered runnable that matches the command line
// arguments. If no runnable matches, the help text is printed.
func (a *Application) ExecuteWithArgs(ctx context.Context, cmdArgs []string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, done := context.WithCancel(ctx)
	defer done()

	go func() {
		<-c
		done()
	}()

	if len(cmdArgs) < 1 {
		fmt.Println("todo print help")
		return
	}

	cmdName := cmdArgs[0]
	if cmdName == "help" {
		if len(cmdArgs) == 1 {
			a.Help()
		} else {
			a.HelpCommand(cmdArgs[1])
		}
		return
	}

	cmd, ok := a.runnables[cmdName]
	if !ok {
		fmt.Fprintf(a.Out, "unknown command: %s\n", cmdName)
		return
	}

	if len(cmdArgs) == 1 {
		cmd.RunCommand(ctx, a.Out)
		return
	}

	parsedArgs, err := parse(strings.Join(cmdArgs[1:], " "))

	if err != nil {
		panic(err)
	}

	t := reflect.TypeOf(cmd)
	v := reflect.ValueOf(cmd)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		argType := t.Field(i)
		flagName := argType.Tag.Get("flag")

		argVal := v.Field(i)

		if !argVal.CanSet() {
			continue
		}

		if flagName == "" {
			continue
		}
		_, hasFlag := parsedArgs[flagName]
		if !hasFlag && argType.Tag.Get("required") == "true" {
			fmt.Fprintf(a.Out, "missing required flag: %s", flagName)
			return
		} else if !hasFlag {
			continue
		}

		parsedArg, ok := parsedArgs[flagName]
		if !ok {
			fmt.Fprintf(a.Out, "could not parse flag %s\n", flagName)
			a.HelpCommand(cmdName)
			return
		}

		switch argType.Type.Kind() {
		case reflect.String:
			v.Field(i).SetString(parsedArg.value)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// convert the string to an int
			intVal, err := strconv.ParseInt(parsedArg.value, 10, 64)
			if err != nil {
				panic(err)
			}

			v.Field(i).SetInt(int64(intVal))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			intVal, err := strconv.ParseUint(parsedArg.value, 10, 64)
			if err != nil {
				panic(err)
			}

			v.Field(i).Set(reflect.ValueOf(intVal))
		case reflect.Bool:
			if parsedArg.value == "f" || parsedArg.value == "false" || parsedArg.value == "0" {
				v.Field(i).SetBool(false)
			} else {
				v.Field(i).SetBool(true)
			}
		default:
			panic(fmt.Sprintf("unsupported type %s", argType.Type.Kind()))
		}
	}

	err = cmd.RunCommand(ctx, a.Out)
	if err != nil {
		panic(err)
	}
}

var cmdNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z:]+$`)

// RegisterCommand adds a runnable to the application that can be run via the CLI.
func (a *Application) RegisterCommand(runnable Runnable, name string, description string) {
	if name == "help" {
		panic("cannot register command named help")
	}

	if !cmdNameRegex.MatchString(name) {
		panic(fmt.Sprintf("invalid command name %s. Command must be alphanumeric and may only contain : special characters", name))
	}

	if len(name) > 20 {
		panic(fmt.Sprintf("command name %s is too long. Command names must be less than 20 characters", name))
	}

	a.runnables[name] = runnable
	a.runnableDescriptions[name] = description
}

func (a *Application) Help() {
	_, _ = a.Out.Write([]byte("usage\n"))

	longestRunnable := 2

	names := make([]string, 0, len(a.runnables))
	for name := range a.runnables {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if len(name) > longestRunnable {
			longestRunnable = len(name)
		}
	}

	for _, name := range names {
		fmt.Fprintf(a.Out, "  %s %s %s\n", name, strings.Repeat(" ", longestRunnable-len(name)), a.runnableDescriptions[name])
	}
}

func (a *Application) HelpCommand(cmdName string) {
	_, _ = a.Out.Write([]byte(fmt.Sprintf("usage for %s\n", cmdName)))

	longestArg := 2

	cmd, ok := a.runnables[cmdName]
	if !ok {
		fmt.Fprintf(a.Out, "unknown command: %s\n", cmdName)
		a.Help()
		return
	}

	t := reflect.TypeOf(cmd)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		argType := t.Field(i)
		flagName := argType.Tag.Get("flag")

		if flagName == "" {
			continue
		}

		if len(flagName) > longestArg {
			longestArg = len(flagName)
		}
	}

	for i := 0; i < t.NumField(); i++ {
		argType := t.Field(i)
		flagName := argType.Tag.Get("flag")
		description := argType.Tag.Get("description")
		// required := argType.Tag.Get("required")

		if flagName == "" {
			continue
		}

		if description == "" {
			description = "no description provided"
		}

		if required := argType.Tag.Get("required"); required == "true" {
			description = fmt.Sprintf("%s (required)", description)
		}
		fmt.Fprintf(a.Out, "  -%s %s %s\n", flagName, strings.Repeat(" ", longestArg-len(flagName)), description)
	}

}
