//go:build wireinject
// +build wireinject

//go:generate go run -mod=mod github.com/google/wire/cmd/wire

package main

import (
	"github.com/google/wire"
	"github.com/spf13/cobra"
	"github.com/tomoya.tokunaga/cli/internal/di"
)

// Application holds the CLI application
type Application struct {
	RootCmd *cobra.Command
}

// Run executes the root command
func (app *Application) Run() error {
	return app.RootCmd.Execute()
}

// InitializeApplication initializes the CLI application with dependency injection
func InitializeApplication() (*Application, error) {
	wire.Build(
		di.CLIProvider,
		wire.Struct(new(Application), "*"),
	)
	return nil, nil
}
