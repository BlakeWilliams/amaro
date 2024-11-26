package components

import (
	"runtime/debug"

	"github.com/blakewilliams/amaro/_template/internal/core"
)

type Err500 struct {
	Environment core.Env
	Error       any
}

func (e *Err500) RenderDetailedError() bool {
	return e.Environment != core.EnvProduction
}

func (e *Err500) ErrorType() string {
	if err, ok := e.Error.(error); ok {
		return err.(error).Error()
	} else {
		return "Unknown error type"
	}
}

func (e *Err500) ErrorDetails() string {
	return string(debug.Stack())
}
