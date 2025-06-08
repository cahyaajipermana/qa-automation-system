package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"qa-automation-system/backend/config"
	"qa-automation-system/backend/routes"
)

func main() {
	// Initialize database
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create screenshots directory if it doesn't exist
	screenshotsDir := "screenshots"
	if err := os.MkdirAll(screenshotsDir, 0755); err != nil {
		log.Fatalf("Failed to create screenshots directory: %v", err)
	}

	// Serve static files from screenshots directory
	router := routes.SetupRouter(db)
	router.Static("/screenshots", "./screenshots")

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Start server
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 