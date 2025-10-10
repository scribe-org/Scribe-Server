// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/api/handlers"
)

// SetupRoutes configures all API routes with versioning.
func SetupRoutes(r *gin.Engine) {
	r.GET("/", handlers.Hello)

	api := r.Group("/api")
	{
		// API v1 routes.
		v1 := api.Group("/v1")
		{
			v1.GET("/data/:lang", handlers.GetLanguageData)
			v1.GET("/data-version/:lang", handlers.GetLanguageVersion)
			v1.GET("/languages", handlers.GetAvailableLanguages)
			v1.GET("/contracts", handlers.GetContracts)
			v1.GET("/language-stats", handlers.GetLanguageStats)
		}
	}
}
