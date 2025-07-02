// SPDX-License-Identifier: GPL-3.0-or-later
package api

import "github.com/gin-gonic/gin"

// SetupCORS adds CORS middleware for API access
func SetupCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET")
		
		c.Next()
	}
}