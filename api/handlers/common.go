// SPDX-License-Identifier: GPL-3.0-or-later

// Package handlers contains HTTP handlers for various API endpoints.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/models"
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

// ServeHome handles the root endpoint.
func ServeHome(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.File("./static/index.html")
}
