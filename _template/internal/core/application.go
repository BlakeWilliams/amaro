package core

import (
	_ "embed"
	"log/slog"
	"os"
	"sync"

	"github.com/blakewilliams/amaro/envy"
)

type Env string

const (
	EnvDevelopment Env = "development"
	EnvTest        Env = "test"
	EnvProduction  Env = "production"
)

// Application holds global application concerns that are shared across
// various components, like the router, jobs, etc.
type Application struct {
	// The environment the application is running in, e.g. "development",
	Environment Env
	Logger      *slog.Logger
}

var once sync.Once

// New returns a new Application instance.
func New(env Env) *Application {
	once.Do(func() {
		if env == EnvProduction {
			return
		}

		err := envy.Load(string(env))
		if err != nil {
			panic(err)
		}
	})

	return &Application{
		Environment: env,
		Logger:      slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}
