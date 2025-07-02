// SPDX-License-Identifier: GPL-3.0-or-later
package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/models"
)

func hello(c *gin.Context) {
	c.String(http.StatusOK, "Hello, I'm Scribe!")
}

// getLanguageData handles GET /data/:lang
func getLanguageData(c *gin.Context) {
	lang := c.Param("lang")
	
	// Validate language code (ISO 639-1)
	if len(lang) != 2 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid language code. Use ISO 639-1 format (e.g., 'en', 'fr')",
		})
		return
	}

	// Mock data - only support 'en' for now - WILL EXPAND ON THIS
	if lang != "en"{
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: fmt.Sprintf("Language '%s' not supported", lang),
		})
		return
	}

	// Return mock response
	response := models.LanguageDataResponse{
		Language: lang,
		Contract: models.Contract{
			Version:   "1.0.0",
			UpdatedAt: time.Now().Format("2025-06-30"),
			Fields: map[string]map[string]string{
				"nouns": {
					"word":   "VARCHAR(255)",
					"plural": "VARCHAR(255)",
					"gender": "VARCHAR(10)",
				},
				"verbs": {
					"infinitive": "VARCHAR(255)",
					"past":       "VARCHAR(255)",
					"present":    "VARCHAR(255)",
				},
				"prepositions": {
					"word": "VARCHAR(255)",
					"type": "VARCHAR(50)",
				},
			},
		},
		Data: map[string]interface{}{
			"nouns": []map[string]interface{}{
				{"word": "cat", "plural": "cats", "gender": "neuter"},
				{"word": "dog", "plural": "dogs", "gender": "neuter"},
				{"word": "house", "plural": "houses", "gender": "neuter"},
			},
			"verbs": []map[string]interface{}{
				{"infinitive": "run", "past": "ran", "present": "runs"},
				{"infinitive": "walk", "past": "walked", "present": "walks"},
				{"infinitive": "eat", "past": "ate", "present": "eats"},
			},
			"prepositions": []map[string]interface{}{
				{"word": "in", "type": "spatial"},
				{"word": "on", "type": "spatial"},
				{"word": "at", "type": "temporal"},
			},
		},
	}

	c.JSON(http.StatusOK, response)
}

// getLanguageVersion handles GET /data-version/:lang
func getLanguageVersion(c *gin.Context) {
	lang := c.Param("lang")
	
	// Validate language code
	if len(lang) != 2 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid language code. Use ISO 639-1 format (e.g., 'en', 'fr')",
		})
		return
	}

	// Mock data - only support 'en' for now (WILL EXPAND ON THIS)
	if lang != "en" {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: fmt.Sprintf("Language '%s' not supported", lang),
		})
		return
	}

	response := models.LanguageVersionResponse{
		Language:                 lang,
		NounsLastModified:       "2025-06-28",
		VerbsLastModified:       "2025-06-30",
		PrepositionsLastModified: "2025-07-01",
	}

	c.JSON(http.StatusOK, response)
}