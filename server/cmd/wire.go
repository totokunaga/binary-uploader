//go:build wireinject
// +build wireinject

//go:generate go run -mod=mod github.com/google/wire/cmd/wire

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/wire"
	"github.com/tomoya.tokunaga/server/internal/di"
	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	"github.com/tomoya.tokunaga/server/internal/interface/api/router"
	"golang.org/x/exp/slog"
)

// Application holds all the components needed for application
type Application struct {
	Config *entity.Config
	Logger *slog.Logger
	Router *router.Router
	Server *http.Server
}

// HTTPServerProvider provides the configured HTTP server
func HTTPServerProvider(config *entity.Config, router *router.Router) *http.Server {
	r := router.Engine()

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: r,
	}
}

// Run starts the application and handles graceful shutdown
func (app *Application) Run() error {
	// Start the server in a goroutine
	go func() {
		app.Logger.Info("starting server", "addr", app.Server.Addr)
		if err := app.Server.ListenAndServe(); err != nil {
			if err.Error() != "http: Server closed" {
				app.Logger.Error("server error", "error", err)
				os.Exit(1)
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	app.Logger.Info("shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server
	if err := app.Server.Shutdown(ctx); err != nil {
		app.Logger.Error("server forced to shutdown", "error", err)
		return err
	}

	app.Logger.Info("server exited gracefully")
	return nil
}

// InitializeApplication initializes all application components
func InitializeApplication() (*Application, error) {
	wire.Build(
		di.ServerProvider,
		HTTPServerProvider,
		wire.Struct(new(Application), "*"),
	)
	return nil, nil
}
