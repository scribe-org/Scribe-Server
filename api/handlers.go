// SPDX-License-Identifier: GPL-3.0-or-later
package api

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func hello(c *gin.Context) {
	c.String(http.StatusOK, "Hello, I'm Scribe!")
}

// HandleRequests handles incoming HTTP requests.
func HandleRequests() {
	r := gin.Default()

	r.GET("/", hello)

	// Setup /packs handler for static files
	fileSystem := viper.GetString("fileSystem")
	log.Printf("Serving files from: %s", fileSystem)

	// Check if the directory exists
	if _, err := os.Stat(fileSystem); os.IsNotExist(err) {
		log.Fatalf("Directory %s does not exist", fileSystem)
	}

	// Serve static files from the /packs route
	r.Static("/packs", fileSystem)

	// Start the server
	hostPort := fmt.Sprintf(":%s", viper.GetString("hostPort"))
	log.Printf("Listening on port %s", hostPort)
	log.Fatal(r.Run(hostPort))
}
