package main

import (
	"log"

	"github.com/codemi-be/golang-browser-mobile/pkg/app"
	"github.com/codemi-be/golang-browser-mobile/pkg/config"
)

func main() {
	// Initialize configuration
	cfg := config.New()

	// Create and run the application
	app := app.New(cfg)

	if err := app.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
