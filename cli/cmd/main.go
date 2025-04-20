package main

import (
	"fmt"
	"os"
)

func main() {
	// Initialize the application with dependency injection
	app, err := InitializeApplication()
	if err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	// Run the application
	if err := app.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
