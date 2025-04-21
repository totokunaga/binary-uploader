package main

import (
	"log"
	"os"
)

func main() {
	// Initialize the application with dependency injection
	app, err := InitializeApplication()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
		os.Exit(1)
	}

	// Run the application
	if err := app.Run(); err != nil {
		app.Logger.Error("Application error", "error", err)
		os.Exit(1)
	}
}
