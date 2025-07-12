// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/database"
	"github.com/spf13/viper"
)

// HandleRequests sets up and starts the server.
func HandleRequests() {
	// Initialize database connection.
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create language_data_versions table if it doesn't exist.
	if err := database.CreateLanguageDataVersionsTable(); err != nil {
		log.Printf("Warning: Failed to create language_data_versions table: %v", err)
	}

	// Set Gin mode based on environment.
	switch viper.GetString("GIN_MODE") {
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode) // fallback
	}

	// Create Gin router with default middleware (logger and recovery).
	r := gin.Default()

	// Add custom middleware.
	r.Use(SetupCORS())

	// Setup API routes.
	SetupRoutes(r)

	// Setup static file serving for existing functionality.
	setupStaticFiles(r)

	// Start the server.
	startServer(r)
}

// setupStaticFiles configures static file serving.
func setupStaticFiles(r *gin.Engine) {
	fileSystem := viper.GetString("fileSystem")
	if fileSystem != "" {
		log.Printf("Serving files from: %s", fileSystem)

		// Check if the directory exists.
		if _, err := os.Stat(fileSystem); os.IsNotExist(err) {
			log.Printf("Warning: Directory %s does not exist, static file serving disabled", fileSystem)
		} else {
			r.Static("/packs", fileSystem)
		}
	}
}

// startServer starts the HTTP server.
func startServer(r *gin.Engine) {
	hostPort := fmt.Sprintf(":%s", viper.GetString("hostPort"))

	// Get available languages for startup message.
	availableLanguages, err := database.GetAvailableLanguages()
	if err != nil {
		log.Printf("Warning: Could not fetch available languages: %v", err)
		availableLanguages = []string{"unknown"}
	}

	log.Printf("👀 Listening on port %s", hostPort)
	log.Printf("🚀 API endpoints available:")
	log.Printf("  ✅ GET /v1/data/:lang - Versioned API")
	log.Printf("  ✅ GET /v1/data-version/:lang - Version check API")
	log.Printf("  ✅ GET /v1/languages - List available languages")
	log.Printf("📊 Available languages: %v", availableLanguages)

	log.Fatal(r.Run(hostPort))
}
