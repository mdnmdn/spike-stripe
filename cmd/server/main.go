package main

import (
	"embed"
	"log"
	"pa11y-go-wrapper/internal/analysis"
	"pa11y-go-wrapper/internal/api"
)

//go:embed frontend
var frontendAssets embed.FS

func main() {
	// Initialize the analysis service
	analysisService := analysis.NewService(100) // Queue size of 100

	// Start the background worker
	worker := analysis.NewWorker(analysisService)
	worker.Start()

	// Create and run the Gin server
	router := api.NewRouter(analysisService, frontendAssets)
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
