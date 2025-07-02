// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// HandleRequests sets up and starts the server
func HandleRequests() {
	// Set Gin mode based on environment
	if viper.GetString("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router with default middleware (logger and recovery)
	r := gin.Default()

	// Add custom middleware
	r.Use(SetupCORS())

	// Setup API routes
	SetupRoutes(r)

	// Setup static file serving for existing functionality
	setupStaticFiles(r)

	// Start the server
	startServer(r)
}

// setupStaticFiles configures static file serving
func setupStaticFiles(r *gin.Engine) {
	fileSystem := viper.GetString("fileSystem")
	if fileSystem != "" {
		log.Printf("Serving files from: %s", fileSystem)

		// Check if the directory exists
		if _, err := os.Stat(fileSystem); os.IsNotExist(err) {
			log.Printf("Warning: Directory %s does not exist, static file serving disabled", fileSystem)
		} else {
			r.Static("/packs", fileSystem)
		}
	}
}

// startServer starts the HTTP server
func startServer(r *gin.Engine) {
	hostPort := fmt.Sprintf(":%s", viper.GetString("hostPort"))
	
	log.Printf("👀 Listening on port %s", hostPort)
	log.Printf("🚀 API endpoints available:")
	log.Printf("  ✅ GET /v1/data/:lang - Versioned API")
	log.Printf("  ✅ GET /v1/data-version/:lang - Versioned API")
	log.Printf("Mock data available for language: en")
	
	log.Fatal(r.Run(hostPort))
}