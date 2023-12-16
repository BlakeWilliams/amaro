// Amaro is an application framework in the form of a library that integrates
// with your application. It's extensible so third-party packages can add
// functionality to your application in addition to first-party features added
// per-project.
package amaro

import (
	"context"
	"fmt"
	"io"
)

type (
	Application struct {
		// The name of the application
		Name string

		// Out is the writer to which output is written. If nil, os.Stdout is used.
		Out io.Writer

		runnables map[string]Runnable
	}

	// Runnable is an interface that can be implemented by any type that
	// wants to be run by the application.
	Runnable interface {
		RunCommand(context.Context) error
	}
)

// NewApplication creates a new application instance.
func NewApplication(name string) *Application {
	return &Application{
		Name:      name,
		runnables: make(map[string]Runnable, 0),
	}
}

// Execute runs the registered runnable that matches the command line arguments.
// If no runnable matches, the help text is printed.
func (a *Application) Execute() {
}

// ExecuteWithArgs runs the registered runnable that matches the command line
// arguments. If no runnable matches, the help text is printed.
func (a *Application) ExecuteWithArgs(cmdArgs []string) {
	if len(cmdArgs) < 2 {
		fmt.Println("todo print help")
		return
	}

	// cmdName := cmdArgs[0]
	// cmd, ok := a.runnables[cmdName]
	// if !ok {
	// 	fmt.Println("todo print help and say command does not exist")
	// 	return
	// }

	// args := make([]string, 0)
	// if len(os.Args) > 2 {
	// 	args = os.Args[2:]
	// }

	// current := 0
	// for {
	// 	// name := args[current]

	// }
}

// Register adds a runnable to the application that can be run via the CLI.
func (a *Application) Register(name string, runnable Runnable) {
	a.runnables[name] = runnable
}

// type arg struct {
// 	name  string
// 	value string
// }

// func normalizeArgs(inputArgs []string) []arg {
// 	args := make([]arg, 0)
// 	input := strings.Join(inputArgs, " ")

// 	i := 0
// 	for {
// 		if i >= len(input) {
// 			break
// 		}

// 		if input[i] == '-' && input[i+1] == '-' {
// 			i += 2
// 			name := ""
// 			for {
// 				if i >= len(input) {
// 					break
// 				}

// 				if input[i] == ' ' {
// 					break
// 				}

// 				name += string(input[i])
// 				i++
// 			}

// 			// safety check
// 			if i >= len(input) {
// 				break
// 			}

// 			// ignore space
// 			if input[i] == '=' {
// 				i++
// 			}

// 			// parse string
// 			value := ""
// 			quoteCount := 0
// 			for {
// 				if i >= len(input) {
// 					break
// 				}

// 				if input[i] == '"' {
// 					quoteCount++
// 					i++
// 					continue
// 				}

// 				// if quoteCount == 2 {
// 				// 	break
// 				// }

// 				value += string(input[i])
// 				i++
// 			}

// 			args = append(args, arg{
// 				name: name,
// 			})
// 		}

// 	}

// 	return args
// }
