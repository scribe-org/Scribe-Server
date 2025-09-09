// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/api/validators"
	"github.com/scribe-org/scribe-server/database"
	"github.com/spf13/viper"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// HandleRequests sets up and starts the server.
func HandleRequests() {
	// Initialize database connection.
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
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

	// Proxies trust security warning.
	trustedProxies := []string{} // Default empty (trusts nothing)
	if os.Getenv("ENV") == "prod" {
		// Trust RFC 1918 private network ranges commonly used by cloud providers and load balancers.
		trustedProxies = []string{
			"10.0.0.0/8", // Class A private range (10.0.0.0 - 10.255.255.255)
			// Note: Used by AWS VPC, Google Cloud, Kubernetes clusters, most cloud providers.
			"172.16.0.0/12", // Class B private range (172.16.0.0 - 172.31.255.255)
			// Note: Used by Docker default bridge networks, some enterprise networks.
			"192.168.0.0/16", // Class C private range (192.168.0.0 - 192.168.255.255)
			// Note: Used by Home routers, small office networks, some internal services.
		}
		log.Printf("🔒 Production mode: trusting proxy networks")
	} else {
		log.Printf("🔒 Development mode: trusting no proxies")
	}

	if err := r.SetTrustedProxies(trustedProxies); err != nil {
		log.Fatal(err)
	}

	// Add custom middleware.
	r.Use(SetupCORS())

	// Setup API routes.
	SetupRoutes(r)

	// Setup Swagger documentation route.
	setupSwagger(r)

	// Setup static file serving for existing functionality.
	setupStaticFiles(r)

	// Start the server.
	startServer(r)
}

// setupSwagger configures the Swagger documentation endpoint.
func setupSwagger(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	log.Printf("📖 API Documentation available at: /swagger/index.html")

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
	r.StaticFile("/favicon.ico", "./static/favicon.ico")
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

	// Initialize cached language validation map.
	validators.InitLanguageValidator(availableLanguages)

	log.Printf("👀 Listening on port %s", hostPort)
	log.Println("🚀 API Endpoints:")
	log.Println("  ✅ GET /api/v1/languages                		- List available languages")
	log.Println("  ✅ GET /api/v1/contracts[?lang_iso=xx]      	- Get contracts (optional language filter)")
	log.Println("  ✅ GET /api/v1/data/:lang_iso       		- Get full language data with schema")
	log.Println("  ✅ GET /api/v1/data-version/:lang_iso 		- Get version info for a language")
	log.Printf("📊 Available languages: %v", availableLanguages)

	log.Fatal(r.Run(hostPort))
}
