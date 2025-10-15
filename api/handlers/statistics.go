// SPDX-License-Identifier: GPL-3.0-or-later

// Package handlers contains HTTP handlers for statistics-related API endpoints.
package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/api/validators"
	"github.com/scribe-org/scribe-server/database"
	"github.com/scribe-org/scribe-server/models"
)

// GetLanguageStats handles GET /language-stats?codes=en,fr.
// @Summary Get statistics for one or multiple languages
// @Description Returns the number of nouns and verbs for the specified language codes.
// @Tags statistics
// @Accept json
// @Produce json
// @Param codes query string false "Comma-separated list of language codes to filter (e.g., "fr,de,es")"
// @Success 200 {array} models.LanguageStatisticsReponse "List of language statistics"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 404 {object} models.ErrorResponse "Language not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/v1/language-stats [get]
func GetLanguageStats(c *gin.Context) {
	codesParam := c.Query("codes")

	if codesParam == "" {
		allStats, err := database.GetAllLanguageStats()
		if err != nil {
			log.Printf("Error fetching all stats: %v", err)
			HandleError(c, http.StatusInternalServerError, "Failed to fetch all language statistics")
			return
		}
		HandleSuccess(c, allStats)
		return
	}

	requestedCodes := strings.Split(codesParam, ",")
	for i := range requestedCodes {
		requestedCodes[i] = strings.TrimSpace(requestedCodes[i])
	}

	availableLanguages, err := database.GetAvailableLanguages()
	if err != nil {
		log.Printf("Error checking available languages: %v", err)
		HandleError(c, http.StatusInternalServerError, "Failed to check available languages")
		return
	}

	supported := make([]string, 0)
	for _, code := range requestedCodes {
		if validators.IsLanguageSupported(code, availableLanguages) {
			supported = append(supported, code)
		}
	}

	statsList := make([]models.LanguageStatisticsReponse, 0)
	for _, code := range supported {
		stat, err := database.GetLanguageStat(code)
		if err != nil {
			log.Printf("Error fetching stats for %s: %v", code, err)
			continue
		}
		if stat == nil {
			continue
		}

		statsList = append(statsList, database.BuildLanguageStatResponse(code, stat))
	}

	if len(statsList) == 0 {
		HandleError(c, http.StatusNotFound, "No statistics found for requested languages")
		return
	}

	HandleSuccess(c, statsList)
}
