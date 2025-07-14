// SPDX-License-Identifier: GPL-3.0-or-later

// Package handlers contains HTTP handlers for various API endpoints.
package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/scribe-org/scribe-server/api/dbqueries"
	"github.com/scribe-org/scribe-server/api/validators"
	"github.com/scribe-org/scribe-server/internal/constants"
	"github.com/scribe-org/scribe-server/models"
)

// GetAvailableLanguages handles GET /languages.
func GetAvailableLanguages(c *gin.Context) {
	languages, err := dbqueries.GetAvailableLanguages()
	if err != nil {
		log.Printf("Error fetching available languages: %v", err)
		HandleError(c, http.StatusInternalServerError, constants.ErrorFetchingLanguages)
		return
	}

	var languageInfos []models.LanguageInfo
	for _, lang := range languages {
		dataTypes, err := dbqueries.GetLanguageDataTypes(lang)
		if err != nil {
			log.Printf("Error fetching data types for %s: %v", lang, err)
			continue
		}

		languageInfos = append(languageInfos, models.LanguageInfo{
			Code:      lang,
			DataTypes: dataTypes,
		})
	}

	HandleSuccess(c, models.AvailableLanguagesResponse{
		Languages: languageInfos,
	})
}

// GetLanguageData handles GET /data/:lang.
func GetLanguageData(c *gin.Context) {
	lang := c.Param("lang")

	// Validate language code.
	if !validators.IsValidLanguageCode(lang) {
		HandleError(c, http.StatusBadRequest, constants.InvalidLanguageCodeError)
		return
	}

	// Check if language exists in database.
	availableLanguages, err := dbqueries.GetAvailableLanguages()
	if err != nil {
		log.Printf("Error checking available languages: %v", err)
		HandleError(c, http.StatusInternalServerError, "Failed to check language availability")
		return
	}

	if !validators.IsLanguageSupported(lang, availableLanguages) {
		HandleError(c, http.StatusNotFound, fmt.Sprintf("Language '%s' not supported", lang))
		return
	}

	// Get data types for the language.
	dataTypes, err := dbqueries.GetLanguageDataTypes(lang)
	if err != nil {
		log.Printf("Error fetching data types for %s: %v", lang, err)
		HandleError(c, http.StatusInternalServerError, "Failed to fetch language data types")
		return
	}

	// Build the response.
	response := models.LanguageDataResponse{
		Language: lang,
		Contract: models.Contract{
			Version:   constants.APIVersion,
			UpdatedAt: time.Now().Format(constants.DateFormat),
			Fields:    make(map[string]map[string]string),
		},
		Data: make(map[string]interface{}),
	}

	// For each data type, get schema and data.
	for _, dataType := range dataTypes {
		// Sanitize inputs before constructing table name.
		sanitizedLang := validators.SanitizeLanguageCode(lang)
		if sanitizedLang == "" {
			log.Printf("Failed to sanitize language code: %s", lang)
			continue
		}

		tableData, err := dbqueries.GetLanguageTableData(lang, dataType)
		if err != nil {
			log.Printf("Error fetching table data for %s/%s: %v", lang, dataType, err)
			continue
		}

		// Add to response.
		if schema, ok := tableData["schema"].(map[string]string); ok {
			response.Contract.Fields[dataType] = schema
		}
		if data, ok := tableData["data"]; ok {
			response.Data[dataType] = data
		}
	}

	HandleSuccess(c, response)
}

// GetLanguageVersion handles GET /data-version/:lang.
func GetLanguageVersion(c *gin.Context) {
	lang := c.Param("lang")

	// Validate language code.
	if !validators.IsValidLanguageCode(lang) {
		HandleError(c, http.StatusBadRequest, constants.InvalidLanguageCodeError)
		return
	}

	// Check if language exists in database.
	availableLanguages, err := dbqueries.GetAvailableLanguages()
	if err != nil {
		log.Printf("Error checking available languages: %v", err)
		HandleError(c, http.StatusInternalServerError, "Failed to check language availability")
		return
	}

	if !validators.IsLanguageSupported(lang, availableLanguages) {
		HandleError(c, http.StatusNotFound, fmt.Sprintf("Language '%s' not supported", lang))
		return
	}

	// Get version information.
	versions, err := dbqueries.GetLanguageVersions(lang)
	if err != nil {
		log.Printf("Error fetching versions for %s: %v", lang, err)
		HandleError(c, http.StatusInternalServerError, constants.ErrorFetchingLanguageVersions)
		return
	}

	response := models.LanguageVersionResponse{
		Language: lang,
		Versions: versions,
	}

	HandleSuccess(c, response)
}
