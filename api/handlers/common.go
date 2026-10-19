// SPDX-License-Identifier: GPL-3.0-or-later

// Package handlers contains HTTP handlers for various API endpoints.
package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/models"
	"gopkg.in/yaml.v3"
)

// HandleError sends a standardized error response.
func HandleError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, models.ErrorResponse{
		Error: message,
	})
}

// HandleSuccess sends a standardized success response.
func HandleSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

// HandleYAMLSuccess sends a standardized success response for the contract file.
// The file name is used when the response is downloaded.
func HandleYAMLSuccess(c *gin.Context, fileName string, data any) {
	out, err := yaml.Marshal(data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to encode YAML")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
	c.Data(http.StatusOK, "application/x-yaml; charset=utf-8", out)
}

// ServeHome handles the root endpoint.
func ServeHome(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.File("./static/index.html")
}
