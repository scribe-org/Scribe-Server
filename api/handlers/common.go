// SPDX-License-Identifier: GPL-3.0-or-later
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
func HandleSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// Hello handles the root endpoint.
func Hello(c *gin.Context) {
	c.String(http.StatusOK, "Hello, I'm Scribe!")
}
