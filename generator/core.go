package generator

func generateApp(d *driver) error {
	d.createFile("internal/core/application.go", `package core

	import (
		"log/slog"
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
		Logger		*slog.Logger
	}
`, nil)

	return nil
}
