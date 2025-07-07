// SPDX-License-Identifier: GPL-3.0-or-later

package api

import "github.com/gin-gonic/gin"

// SetupRoutes configures all API routes with versioning
func SetupRoutes(r *gin.Engine) {
	r.GET("/", hello)

	api := r.Group("/api")
	{
		// API v1 routes
		v1 := api.Group("/v1")
		{
			v1.GET("/data/:lang", getLanguageData)
			v1.GET("/data-version/:lang", getLanguageVersion)
			v1.GET("/languages", getAvailableLanguages)
		}
	}
}
