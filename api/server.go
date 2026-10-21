// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/api/validators"
	"github.com/scribe-org/scribe-server/database"
	"github.com/spf13/viper"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// MARK: Server Initialization

// Function variables allow mocking the database, server, file paths and fatal errors in unit tests.
var (
	initDatabase          = database.InitDatabase
	getAvailableLanguages = database.GetAvailableLanguages
	runServer             = func(r *gin.Engine, addr string) error { return r.Run(addr) }
	resolveAbsPath        = filepath.Abs
	fatalf                = log.Fatalf
)

// HandleRequests sets up and starts the server.
func HandleRequests() {
	viper.AutomaticEnv()

	// Initialize database connection.
	if err := initDatabase(); err != nil {
		fatalf("Failed to initialize database: %v", err)
		return
	}

	r, err := newRouter()
	if err != nil {
		fatalf("Failed to set up the router: %v", err)
		return
	}

	// Start the server.
	startServer(r)
}

// newRouter creates the Gin router with its middleware, API routes, documentation and static files.
func newRouter() (*gin.Engine, error) {
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

	if err := r.SetTrustedProxies(trustedProxies()); err != nil {
		return nil, err
	}

	// Add custom middleware.
	r.Use(SetupCORS())

	// Setup API routes.
	SetupRoutes(r)

	// Setup Swagger documentation route.
	setupSwagger(r)

	// Setup static file serving for existing functionality.
	setupStaticFiles(r)

	return r, nil
}

// MARK: Proxy Configuration

// prodTrustedProxies are the RFC 1918 private network ranges commonly used by cloud providers and load balancers.
var prodTrustedProxies = []string{
	"10.0.0.0/8", // Class A private range (10.0.0.0 - 10.255.255.255)
	// Note: Used by AWS VPC, Google Cloud, Kubernetes clusters, most cloud providers.
	"172.16.0.0/12", // Class B private range (172.16.0.0 - 172.31.255.255)
	// Note: Used by Docker default bridge networks, some enterprise networks.
	"192.168.0.0/16", // Class C private range (192.168.0.0 - 192.168.255.255)
	// Note: Used by Home routers, small office networks, some internal services.
}

// trustedProxies returns the proxy networks to trust: private networks in production and none otherwise.
func trustedProxies() []string {
	if os.Getenv("ENV") == "prod" {
		log.Printf("🔒 Production mode: trusting proxy networks")
		return prodTrustedProxies
	}

	log.Printf("🔒 Development mode: trusting no proxies")
	return []string{}
}

// MARK: Swagger Documentation

// setupSwagger configures the Swagger documentation endpoint.
func setupSwagger(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler()))
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler()))
	log.Printf("📖 API Documentation available at: /swagger/index.html")

}

// MARK: Static File Handling

// setupStaticFiles configures static file serving.
func setupStaticFiles(r *gin.Engine) {
	fileSystem := viper.GetString("fileSystem")
	if fileSystem == "" {
		fileSystem = "./packs"
	}

	absPath, err := resolveAbsPath(fileSystem)
	if err != nil {
		fatalf("Error resolving absolute path for %s: %v", fileSystem, err)
		return
	}

	log.Printf("Serving files from: %s", absPath)
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Printf("Warning: Directory %s does not exist, static file serving disabled", absPath)
		return
	}

	sqlitePath := filepath.Join(absPath, "sqlite")
	log.Printf("SQLite directory path: %s", sqlitePath)

	// Single unified handler for all /packs/* routes.
	r.GET("/packs/*filepath", func(c *gin.Context) {
		path := c.Param("filepath")

		// Prevent directory traversal, which would also show whether directories outside of packs exist.
		if slices.Contains(strings.Split(path, "/"), "..") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path"})
			return
		}

		if path == "/sqlite/list" {
			files, err := os.ReadDir(sqlitePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read directory"})
				return
			}
			sqliteFiles := []string{}
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".sqlite") {
					sqliteFiles = append(sqliteFiles, file.Name())
				}
			}
			c.JSON(http.StatusOK, sqliteFiles)
			return
		}

		if strings.HasPrefix(path, "/sqlite/") && strings.HasSuffix(path, ".sqlite") {
			filename := strings.TrimPrefix(path, "/sqlite/")
			// Prevent directory traversal.
			if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filename"})
				return
			}

			filePath := filepath.Join(sqlitePath, filename)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
				return
			}
			// Set headers to force download.
			c.Header("Content-Description", "File Transfer")
			c.Header("Content-Transfer-Encoding", "binary")
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
			c.Header("Content-Type", "application/x-sqlite3")
			c.File(filePath)
			return
		}

		fullPath := filepath.Join(absPath, path)

		if info, err := os.Stat(fullPath); err == nil && info.IsDir() && !strings.HasSuffix(c.Request.URL.Path, "/") {
			c.Redirect(http.StatusMovedPermanently, c.Request.URL.Path+"/")
			return
		}

		fs := http.FileServer(http.Dir(absPath))
		http.StripPrefix("/packs/", fs).ServeHTTP(c.Writer, c.Request)
	})

	faviconPath := filepath.Join("static", "favicon.ico")
	if _, err := os.Stat(faviconPath); err == nil {
		r.StaticFile("/favicon.ico", faviconPath)
	}

	logoPath := filepath.Join("static", "images", "ScribeServerLogoWhite.png")
	if _, err := os.Stat(logoPath); err == nil {
		r.StaticFile("/images/ScribeServerLogoWhite.png", logoPath)
	}
}

// MARK: Server Startup

// startServer starts the HTTP server.
func startServer(r *gin.Engine) {
	hostPort := fmt.Sprintf(":%s", viper.GetString("hostPort"))

	// Get available languages for startup message.
	availableLanguages, err := getAvailableLanguages()
	if err != nil {
		log.Printf("Warning: Could not fetch available languages: %v", err)
		availableLanguages = []string{"unknown"}
	}

	// Initialize cached language validation map.
	validators.InitLanguageValidator(availableLanguages)

	log.Printf("👀 Listening on port %s", hostPort)
	log.Println("🚀 API Endpoints:")
	log.Println("  ✅ GET /api/v1/languages                				- List available languages")
	log.Println("  ✅ GET /api/v1/contracts[?lang_iso=xx]      			- Get contracts (optional language filter)")
	log.Println("  ✅ GET /api/v1/data/:lang_iso       				- Get full language data with schema")
	log.Println("  ✅ GET /api/v1/data-version/:lang_iso 				- Get version info for a language")
	log.Println("  ✅ GET /api/v1/language-stats?codes=fr,de         		- Get statistics for all or selected languages")
	log.Println("  ✅ GET /api/v1/translations?source_lang=es&target_lang=en  	- Get translation data of target from source")
	log.Printf("📊 Available languages: %v", availableLanguages)

	if err := runServer(r, hostPort); err != nil {
		fatalf("Failed to run the server: %v", err)
	}
}
