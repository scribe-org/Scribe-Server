// SPDX-License-Identifier: GPL-3.0-or-later

package api

import "github.com/gin-gonic/gin"

// SetupRoutes configures all API routes with versioning
func SetupRoutes(r *gin.Engine) {
	r.GET("/", hello)

	// API v1 routes
	v1 := r.Group("/v1")
	{
		v1.GET("/data/:lang", getLanguageData)
		v1.GET("/data-version/:lang", getLanguageVersion)
	}
}
